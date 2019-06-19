package app

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	kappAppLabelKey = "kapp.k14s.io/app"
)

type RecordedApp struct {
	name   string
	nsName string

	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface

	memoizedMeta *AppMeta
}

var _ App = &RecordedApp{}

func (a *RecordedApp) Name() string { return a.name }

func (a *RecordedApp) LabelSelector() (labels.Selector, error) {
	app, err := a.labeledApp()
	if err != nil {
		return nil, err
	}

	return app.LabelSelector()
}

func (a *RecordedApp) CreateOrUpdate(labels map[string]string) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.nsName,
			Labels: map[string]string{
				kappIsAppLabelKey: kappIsAppLabelValue,
			},
		},
		Data: AppMeta{
			LabelKey:   kappAppLabelKey,
			LabelValue: fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		}.AsData(),
	}

	err := a.mergeAppUpdates(configMap, labels)
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Create(configMap)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			existingConfigMap, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(a.name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("Getting app: %s", err)
			}

			err = a.mergeAppUpdates(existingConfigMap, labels)
			if err != nil {
				return err
			}

			_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(existingConfigMap)
			if err != nil {
				return fmt.Errorf("Updating app: %s", err)
			}

			return nil
		}

		return fmt.Errorf("Creating app: %s", err)
	}

	return nil
}

func (a *RecordedApp) mergeAppUpdates(cm *corev1.ConfigMap, labels map[string]string) error {
	for key, val := range labels {
		if prevVal, found := cm.ObjectMeta.Labels[key]; found {
			if prevVal != val {
				return fmt.Errorf("Expected label '%s' value to remain same", key)
			}
		}
		cm.ObjectMeta.Labels[key] = val
	}

	return nil
}

func (a *RecordedApp) Exists() (bool, error) {
	_, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(a.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("Getting app: %s", err)
	}

	return true, nil
}

func (a *RecordedApp) Delete() error {
	app, err := a.labeledApp()
	if err != nil {
		return err
	}

	err = NewRecordedAppChanges(a.nsName, a.name, a.coreClient).DeleteAll()
	if err != nil {
		return fmt.Errorf("Deleting app changes: %s", err)
	}

	err = app.Delete()
	if err != nil {
		return err
	}

	// TODO verify that there no more resources
	err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(a.name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting app: %s", err)
	}

	return nil
}

func (a *RecordedApp) labeledApp() (*LabeledApp, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	sel := labels.Set(meta.Labels()).AsSelector()

	return &LabeledApp{sel, a.coreClient, a.dynamicClient}, nil
}

func (a *RecordedApp) Meta() (AppMeta, error) { return a.meta() }

func (a *RecordedApp) meta() (AppMeta, error) {
	if a.memoizedMeta != nil {
		// set if bulk read on initialization
		return *a.memoizedMeta, nil
	}

	app, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(a.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return AppMeta{}, fmt.Errorf("App '%s' (namespace: %s) does not exist: %s", a.name, a.nsName, err)
		}
		return AppMeta{}, fmt.Errorf("Getting app: %s", err)
	}

	return NewAppMetaFromData(app.Data), nil
}

func (a *RecordedApp) Changes() ([]Change, error) {
	return NewRecordedAppChanges(a.nsName, a.name, a.coreClient).List()
}

func (a *RecordedApp) LastChange() (Change, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	if len(meta.LastChangeName) == 0 {
		return nil, nil
	}

	change := &ChangeImpl{
		name:       meta.LastChangeName,
		nsName:     a.nsName,
		coreClient: a.coreClient,
		meta:       meta.LastChange,
	}

	return change, nil
}

func (a *RecordedApp) BeginChange(meta ChangeMeta) (Change, error) {
	change, err := NewRecordedAppChanges(a.nsName, a.name, a.coreClient).Begin(meta)
	if err != nil {
		return nil, err
	}

	memoizingChange := appTrackingChange{change, a}

	err = memoizingChange.syncOnApp()
	if err != nil {
		_ = change.Fail()
		return nil, err
	}

	return memoizingChange, nil
}

func (a *RecordedApp) update(doFunc func(*AppMeta)) error {
	change, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(a.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Getting app: %s", err)
	}

	meta := NewAppMetaFromData(change.Data)
	doFunc(&meta)

	change.Data = meta.AsData()

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(change)
	if err != nil {
		return fmt.Errorf("Updating app: %s", err)
	}

	return nil
}

type appTrackingChange struct {
	change *ChangeImpl
	app    *RecordedApp
}

var _ Change = appTrackingChange{}

func (c appTrackingChange) Name() string     { return c.change.Name() }
func (c appTrackingChange) Meta() ChangeMeta { return c.change.meta }

func (c appTrackingChange) Fail() error {
	err := c.change.Fail()
	if err != nil {
		return err
	}

	_ = c.syncOnApp()

	return err
}

func (c appTrackingChange) Succeed() error {
	err := c.change.Succeed()
	if err != nil {
		return err
	}

	_ = c.syncOnApp()

	return err
}

func (c appTrackingChange) syncOnApp() error {
	return c.app.update(func(meta *AppMeta) {
		meta.LastChangeName = c.change.Name()
		meta.LastChange = c.change.meta
	})
}
