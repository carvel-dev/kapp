// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

const (
	kappAppLabelKey = "kapp.k14s.io/app"
)

type RecordedApp struct {
	name              string
	fqName            string
	nsName            string
	isMigrated        bool
	creationTimestamp time.Time

	coreClient             kubernetes.Interface
	identifiedResources    ctlres.IdentifiedResources
	appInDiffNsHintMsgFunc func(string) string

	memoizedMeta *Meta
	logger       logger.Logger
}

var _ App = &RecordedApp{}

func (a *RecordedApp) Name() string      { return a.name }
func (a *RecordedApp) Namespace() string { return a.nsName }

func (a *RecordedApp) CreationTimestamp() time.Time { return a.creationTimestamp }

func (a *RecordedApp) Description() string {
	return fmt.Sprintf("app '%s' namespace: %s", a.name, a.nsName)
}

func (a *RecordedApp) LabelSelector() (labels.Selector, error) {
	app, err := a.labeledApp()
	if err != nil {
		return nil, err
	}

	return app.LabelSelector()
}

func (a *RecordedApp) UsedGVs() ([]schema.GroupVersion, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	return meta.UsedGVs, nil
}

func (a *RecordedApp) usedGKs() (*[]schema.GroupKind, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	return meta.UsedGKs, nil
}

func (a *RecordedApp) UsedGKs() (*[]schema.GroupKind, error) { return a.usedGKs() }

func (a *RecordedApp) UpdateUsedGVsAndGKs(gvs []schema.GroupVersion, gks []schema.GroupKind) error {
	gvsByGV := map[schema.GroupVersion]struct{}{}
	var uniqGVs []schema.GroupVersion

	for _, gv := range gvs {
		if _, found := gvsByGV[gv]; !found {
			gvsByGV[gv] = struct{}{}
			uniqGVs = append(uniqGVs, gv)
		}
	}

	gksByGK := map[schema.GroupKind]struct{}{}
	var uniqGKs []schema.GroupKind

	for _, gk := range gks {
		if _, found := gksByGK[gk]; !found {
			gksByGK[gk] = struct{}{}
			uniqGKs = append(uniqGKs, gk)
		}
	}

	sort.Slice(uniqGVs, func(i int, j int) bool {
		return uniqGVs[i].Group+uniqGVs[i].Version < uniqGVs[j].Group+uniqGVs[j].Version
	})

	sort.Slice(uniqGKs, func(i int, j int) bool {
		return uniqGKs[i].Group+uniqGKs[i].Kind < uniqGKs[j].Group+uniqGKs[j].Kind
	})

	return a.update(func(meta *Meta) {
		meta.UsedGVs = uniqGVs
		meta.UsedGKs = &uniqGKs
	})
}

func (a *RecordedApp) CreateOrUpdate(labels map[string]string) error {
	defer a.logger.DebugFunc("CreateOrUpdate").Finish()

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.nsName,
			Labels: map[string]string{
				KappIsAppLabelKey: kappIsAppLabelValue,
			},
		},
		Data: Meta{
			LabelKey:   kappAppLabelKey,
			LabelValue: fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			UsedGKs:    &[]schema.GroupKind{},
		}.AsData(),
	}

	if a.isMigrationEnabled() {
		migratedAnnotation := map[string]string{KappIsConfigmapMigratedAnnotationKey: KappIsConfigmapMigratedAnnotationValue}

		configMapWithSuffix, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.fqName, metav1.GetOptions{})
		if err == nil {
			if a.isKappApp(configMapWithSuffix) {
				err = a.mergeAppUpdates(configMapWithSuffix, labels)
				if err != nil {
					return err
				}

				_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), configMapWithSuffix, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating app: %s", err)
				}

				return nil
			}
		} else if !errors.IsNotFound(err) {
			// return if error is anything other than configmap not found
			return fmt.Errorf("Getting app: %s", err)
		}

		configmapWithoutSuffix, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.name, metav1.GetOptions{})
		if err == nil {
			if a.isKappApp(configmapWithoutSuffix) {
				err = a.mergeAppAnnotationUpdates(configmapWithoutSuffix, migratedAnnotation)
				if err != nil {
					return err
				}

				return a.migrate(configmapWithoutSuffix, labels)
			}
		} else if !errors.IsNotFound(err) {
			// return if error is anything other than configmap not found
			return fmt.Errorf("Getting app: %s", err)
		}

		err = a.mergeAppAnnotationUpdates(configMap, migratedAnnotation)
		if err != nil {
			return err
		}

		configMap.ObjectMeta.Name = a.fqName
	}

	return a.createOrUpdate(configMap, labels)
}

func (a *RecordedApp) createOrUpdate(c *corev1.ConfigMap, labels map[string]string) error {
	err := a.mergeAppUpdates(c, labels)
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Create(context.TODO(), c, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			existingConfigMap, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), c.GetObjectMeta().GetName(), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("Getting app: %s", err)
			}

			err = a.mergeAppUpdates(existingConfigMap, labels)
			if err != nil {
				return err
			}

			_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), existingConfigMap, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("Updating app: %s", err)
			}

			return nil
		}

		return fmt.Errorf("Creating app: %s", err)
	}

	return nil
}

func (a *RecordedApp) migrate(c *corev1.ConfigMap, labels map[string]string) error {
	err := a.renameConfigMap(c, a.fqName, a.nsName)
	if err != nil {
		return err
	}

	err = a.mergeAppUpdates(c, labels)
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), c, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Updating app: %s", err)
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

func (a *RecordedApp) mergeAppAnnotationUpdates(cm *corev1.ConfigMap, annotations map[string]string) error {
	if cm.ObjectMeta.Annotations == nil && len(annotations) > 0 {
		cm.ObjectMeta.Annotations = map[string]string{}
	}

	for key, val := range annotations {
		if prevVal, found := cm.ObjectMeta.Annotations[key]; found {
			if prevVal != val {
				return fmt.Errorf("Expected annotation '%s' value to remain same", key)
			}
		}
		cm.ObjectMeta.Annotations[key] = val
	}

	return nil
}

func (a *RecordedApp) Exists() (bool, string, error) {
	if a.isMigrationEnabled() {
		_, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.fqName, metav1.GetOptions{})
		if err == nil {
			a.isMigrated = true
			return true, "", nil
		} else if !errors.IsNotFound(err) {
			// return if error is anything other than configmap not found
			return false, "", fmt.Errorf("Getting app: %s", err)
		}
	}

	_, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			desc := fmt.Sprintf("App '%s' (namespace: %s) does not exist%s",
				a.name, a.nsName, a.appInDiffNsHintMsgFunc(a.name))
			return false, desc, nil
		}
		return false, "", fmt.Errorf("Getting app: %s", err)
	}

	return true, "", nil
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

	name := a.name
	if a.isMigrated {
		name = a.fqName
	}

	err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting app: %s", err)
	}

	return nil
}

func (a *RecordedApp) Rename(newName string, newNamespace string) error {
	name := a.name
	if a.isMigrated {
		name = a.fqName
	}

	app, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("App '%s' (namespace: %s) does not exist: %s%s",
				a.name, a.nsName, err, a.appInDiffNsHintMsgFunc(name))
		}

		return fmt.Errorf("Getting app: %s", err)
	}

	// use fully qualified name if app had been previously migrated
	if a.isMigrated || a.isMigrationEnabled() {
		a.mergeAppAnnotationUpdates(app, map[string]string{KappIsConfigmapMigratedAnnotationKey: KappIsConfigmapMigratedAnnotationValue})
		return a.renameConfigMap(app, newName+AppSuffix, newNamespace)
	}

	return a.renameConfigMap(app, newName, newNamespace)
}

func (a *RecordedApp) renameConfigMap(app *corev1.ConfigMap, name, ns string) error {
	oldName := app.Name

	// Clear out all existing meta fields
	app.ObjectMeta = metav1.ObjectMeta{
		Name:        name,
		Namespace:   ns,
		Labels:      app.ObjectMeta.Labels,
		Annotations: app.ObjectMeta.Annotations,
	}

	_, err := a.coreClient.CoreV1().ConfigMaps(ns).Create(context.TODO(), app, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Creating app: %s", err)
	}

	err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(context.TODO(), oldName, metav1.DeleteOptions{})
	if err != nil {
		// TODO Do not clean up new config map as there is no gurantee it can be deleted either
		return fmt.Errorf("Deleting app: %s", err)
	}

	// TODO deal with app history somehow?

	return nil
}

func (a *RecordedApp) labeledApp() (*LabeledApp, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	sel := labels.Set(meta.Labels()).AsSelector()

	return &LabeledApp{sel, a.identifiedResources}, nil
}

func (a *RecordedApp) isKappApp(app *corev1.ConfigMap) bool {
	_, ok := app.Labels[KappIsAppLabelKey]
	return ok
}

func (a *RecordedApp) isMigrationEnabled() bool {
	return strings.ToLower(os.Getenv("KAPP_FQ_CONFIGMAP_NAMES")) == "true"
}

func (a *RecordedApp) Meta() (Meta, error) { return a.meta() }

func (a *RecordedApp) setMeta(app corev1.ConfigMap) (Meta, error) {
	meta, err := NewAppMetaFromData(app.Data)
	if err != nil {
		errMsg := "App '%s' (namespace: %s) backed by ConfigMap '%s' did not contain parseable app metadata: %s"
		hintText := " (hint: ConfigMap was overriden by another user?)"

		if a.isMigrated {
			return Meta{}, fmt.Errorf(errMsg+hintText, a.name, a.nsName, a.fqName, err)
		}
		return Meta{}, fmt.Errorf(errMsg+hintText, a.name, a.nsName, a.name, err)
	}

	a.memoizedMeta = &meta

	return meta, nil
}

func (a *RecordedApp) meta() (Meta, error) {
	if a.memoizedMeta != nil {
		// set if bulk read on initialization
		return *a.memoizedMeta, nil
	}

	if a.isMigrationEnabled() {
		app, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.fqName, metav1.GetOptions{})
		if err == nil {
			a.isMigrated = true
			return a.setMeta(*app)
		} else if !errors.IsNotFound(err) {
			// return if error is anything other than configmap not found
			return Meta{}, fmt.Errorf("Getting app: %s", err)
		}
	}

	app, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), a.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return Meta{}, fmt.Errorf("App '%s' (namespace: %s) does not exist: %s%s",
				a.name, a.nsName, err, a.appInDiffNsHintMsgFunc(a.fqName))
		}
		return Meta{}, fmt.Errorf("Getting app: %s", err)
	}

	return a.setMeta(*app)
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

func (a *RecordedApp) update(doFunc func(*Meta)) error {
	name := a.name
	if a.isMigrated {
		name = a.fqName
	}

	change, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Getting app: %s", err)
	}

	meta, err := NewAppMetaFromData(change.Data)
	if err != nil {
		return err
	}

	doFunc(&meta)

	change.Data = meta.AsData()

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), change, metav1.UpdateOptions{})
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

func (c appTrackingChange) Delete() error {
	return c.change.Delete()
}

func (c appTrackingChange) syncOnApp() error {
	return c.app.update(func(meta *Meta) {
		meta.LastChangeName = c.change.Name()
		meta.LastChange = c.change.meta
	})
}
