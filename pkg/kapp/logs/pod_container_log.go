// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type PodContainerLog struct {
	pod        corev1.Pod
	container  string
	podsClient typedcorev1.PodInterface

	tag  string
	opts PodLogOpts
}

func NewPodContainerLog(
	pod corev1.Pod,
	container string,
	podsClient typedcorev1.PodInterface,
	tag string,
	opts PodLogOpts,
) PodContainerLog {
	return PodContainerLog{
		pod:        pod,
		container:  container,
		podsClient: podsClient,

		tag:  tag,
		opts: opts,
	}
}

func (l PodContainerLog) Tail(ui ui.UI, cancelCh chan struct{}) error {
	linePrefix := ""
	if len(l.opts.LinePrefix) > 0 {
		linePrefix = l.opts.LinePrefix + " | "
	}

	for {
		err := l.StartTail(ui, cancelCh)
		if err == io.EOF {
			if l.opts.Follow {
				ui.BeginLinef("%s# container stopped '%s' logs\n", linePrefix, l.tag)
				// Making it 1sec instead of 500ms as some times for initContainers, it fetches the older stream if we go by 500ms.
				time.Sleep(1 * time.Second)
				continue
			}
			ui.BeginLinef("%s# ending tailing '%s' logs\n", linePrefix, l.tag)
			return nil

		}
		if err != nil {
			return err
		}
	}
}

func (l PodContainerLog) StartTail(ui ui.UI, cancelCh chan struct{}) error {
	var streamCanceled atomic.Value

	linePrefix := ""
	if len(l.opts.LinePrefix) > 0 {
		linePrefix = l.opts.LinePrefix + " | "
	}

	stream, err := l.obtainStream(ui, linePrefix, cancelCh)
	if err != nil {
		return err
	}

	if stream == nil {
		return nil
	}

	defer stream.Close()

	go func() {
		<-cancelCh
		streamCanceled.Store(true)
		stream.Close()
	}()

	reader := bufio.NewReader(stream)

	ui.BeginLinef("%s# starting tailing '%s' logs\n", linePrefix, l.tag)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			typedCanceled, ok := streamCanceled.Load().(bool)
			if ok && typedCanceled {
				return nil // ignore error if stream was canceled
			}
			return err
		}

		if l.opts.ContainerTag {
			ui.PrintBlock([]byte(fmt.Sprintf("%s%s | %s", linePrefix, l.tag, line)))
		} else {
			ui.PrintBlock([]byte(fmt.Sprintf("%s%s", linePrefix, line)))
		}
	}
}

func (l PodContainerLog) obtainStream(ui ui.UI, linePrefix string, cancelCh chan struct{}) (io.ReadCloser, error) {
	var (
		isWaitingMsgPrintedOnce bool
		err                     error
	)
	for {
		// TODO infinite retry
		// It appears that GetLogs will successfully return log stream
		// almost immediately after pod has been created; however,
		// returned log stream will not carry any data, even after containers have started.
		// Wait for pod object to have container status fields initialized
		// since that appears to make GetLogs call return actual log stream.
		if l.readyToGetLogs() {
			logs := l.podsClient.GetLogs(l.pod.Name, &corev1.PodLogOptions{
				Follow:    l.opts.Follow,
				TailLines: l.opts.Lines,
				Container: l.container,
				// TODO other options
			})

			stream, err := logs.Stream(context.TODO())
			if err == nil {
				return stream, nil
			}
		}

		if !isWaitingMsgPrintedOnce {
			ui.BeginLinef("%s# waiting for '%s' logs to become available...\n", linePrefix, l.tag)
			isWaitingMsgPrintedOnce = true
			if !l.opts.Follow {
				if err == nil {
					err = fmt.Errorf("Container %s not in Ready state", l.container)
				}
				return nil, err
			}
		}

		select {
		case <-cancelCh:
			return nil, nil
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (l PodContainerLog) readyToGetLogs() bool {
	pod, err := l.podsClient.Get(context.TODO(), l.pod.Name, metav1.GetOptions{})
	if err != nil {
		return false
	}
	containerStatuses := pod.Status.ContainerStatuses
	initContainerStatuses := pod.Status.InitContainerStatuses
	for _, containerStatus := range containerStatuses {
		if l.container != containerStatus.Name {
			continue
		}
		if containerStatus.State.Running != nil {
			return true
		}
	}
	for _, initContainerStatus := range initContainerStatuses {
		if l.container != initContainerStatus.Name {
			continue
		}
		if initContainerStatus.State.Running != nil {
			return true
		}
	}
	return false
}
