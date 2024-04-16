// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"carvel.dev/kapp/pkg/kapp/logger"
	ctlres "carvel.dev/kapp/pkg/kapp/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

const (
	kappAppLabelKey                        = "kapp.k14s.io/app"
	KappIsConfigmapMigratedAnnotationKey   = "kapp.k14s.io/is-configmap-migrated"
	KappIsConfigmapMigratedAnnotationValue = ""
	AppSuffix                              = ".apps.k14s.io"
	KappAppChangesUseAppLabelAnnotationKey = "kapp.k14s.io/app-changes-use-app-label"
)

type RecordedApp struct {
	name                  string
	nsName                string
	isMigrated            bool
	creationTimestamp     time.Time
	appChangesUseAppLabel bool

	coreClient             kubernetes.Interface
	identifiedResources    ctlres.IdentifiedResources
	appInDiffNsHintMsgFunc func(string) string

	memoizedMeta *Meta
	logger       logger.Logger
}

func NewRecordedApp(name, nsName string, creationTimestamp time.Time, coreClient kubernetes.Interface,
	identifiedResources ctlres.IdentifiedResources, appInDiffNsHintMsgFunc func(string) string, logger logger.Logger) *RecordedApp {

	// Always trim suffix, even if user added it manually (to avoid double migration)
	return &RecordedApp{strings.TrimSuffix(name, AppSuffix), nsName, false, creationTimestamp, false, coreClient, identifiedResources, appInDiffNsHintMsgFunc,
		nil, logger.NewPrefixed("RecordedApp")}
}

var _ App = &RecordedApp{}

func (a *RecordedApp) Name() string      { return a.name }
func (a *RecordedApp) Namespace() string { return a.nsName }

func (a *RecordedApp) fqName() string { return a.name + AppSuffix }

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

func (a *RecordedApp) CreateOrUpdate(prevAppName string, labels map[string]string, isDiffRun bool) (bool, error) {
	defer a.logger.DebugFunc("CreateOrUpdate").Finish()

	app, foundMigratedApp, err := a.find(a.fqName())
	if err != nil {
		return false, err
	}
	if foundMigratedApp {
		a.isMigrated = true
		if isDiffRun {
			return false, nil
		}
		return false, a.updateApp(app, labels)
	}

	app, foundNonMigratedApp, err := a.find(a.name)
	if err != nil {
		return false, err
	}
	if foundNonMigratedApp {
		if isDiffRun {
			return false, nil
		}
		if a.isMigrationEnabled() {
			return false, a.migrate(app, labels, a.fqName())
		}
		return false, a.updateApp(app, labels)
	}

	if prevAppName == "" {
		return true, a.create(labels, isDiffRun)
	}

	app, foundMigratedPrevApp, err := a.find(prevAppName + AppSuffix)
	if err != nil {
		return false, err
	}
	if foundMigratedPrevApp {
		a.isMigrated = true
		if isDiffRun {
			return false, nil
		}
		return false, a.renameConfigMap(app, a.fqName(), a.nsName)
	}

	app, foundNonMigratedPrevApp, err := a.find(prevAppName)
	if err != nil {
		return false, err
	}
	if foundNonMigratedPrevApp {
		if isDiffRun {
			return false, nil
		}
		if a.isMigrationEnabled() {
			return false, a.migrate(app, labels, a.fqName())
		}
		return false, a.renameConfigMap(app, a.name, a.nsName)
	}

	return true, a.create(labels, isDiffRun)
}

func (a *RecordedApp) find(name string) (*corev1.ConfigMap, bool, error) {
	cm, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("Getting app: %w", err)
	}
	return cm, true, nil
}

func (a *RecordedApp) create(labels map[string]string, isDiffRun bool) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.name,
			Namespace: a.nsName,
			Labels: map[string]string{
				KappIsAppLabelKey: kappIsAppLabelValue,
			},
			// Use app label as app change label for all new apps
			Annotations: map[string]string{
				KappAppChangesUseAppLabelAnnotationKey: "",
			},
		},
		Data: Meta{
			LabelKey:   kappAppLabelKey,
			LabelValue: fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
			UsedGKs:    &[]schema.GroupKind{},
		}.AsData(),
	}

	if a.isMigrationEnabled() {
		configMap.ObjectMeta.Name = a.fqName()
		a.isMigrated = true

		err := a.mergeAppAnnotationUpdates(configMap, map[string]string{KappIsConfigmapMigratedAnnotationKey: KappIsConfigmapMigratedAnnotationValue})
		if err != nil {
			return err
		}
	}

	err := a.mergeAppUpdates(configMap, labels)
	if err != nil {
		return err
	}

	if isDiffRun {
		a.setMeta(*configMap)
		return nil
	}

	app, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Create(context.TODO(), configMap, metav1.CreateOptions{})
	a.setMeta(*app)

	return err
}

func (a *RecordedApp) updateApp(existingConfigMap *corev1.ConfigMap, labels map[string]string) error {
	err := a.mergeAppUpdates(existingConfigMap, labels)
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), existingConfigMap, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Updating app: %w", err)
	}

	return nil
}

func (a *RecordedApp) migrate(c *corev1.ConfigMap, labels map[string]string, newName string) error {
	err := a.renameConfigMap(c, newName, a.nsName)
	if err != nil {
		return err
	}

	err = a.mergeAppUpdates(c, labels)
	if err != nil {
		return err
	}

	err = a.mergeAppAnnotationUpdates(c, map[string]string{KappIsConfigmapMigratedAnnotationKey: KappIsConfigmapMigratedAnnotationValue})
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), c, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Updating app: %w", err)
	}

	a.isMigrated = true

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
	_, foundMigratedApp, err := a.find(a.fqName())
	if err != nil {
		return false, "", err
	}

	if foundMigratedApp {
		a.isMigrated = true
		return true, "", nil
	}

	_, foundNonMigratedApp, err := a.find(a.name)
	if err != nil {
		return false, "", err
	}

	if !foundNonMigratedApp {
		desc := fmt.Sprintf("App '%s' (namespace: %s) does not exist%s",
			a.name, a.nsName, a.appInDiffNsHintMsgFunc(a.name))
		return false, desc, nil
	}

	return true, "", nil
}

func (a *RecordedApp) Delete() error {
	app, err := a.labeledApp()
	if err != nil {
		return err
	}

	meta, err := a.meta()
	if err != nil {
		return err
	}

	err = NewRecordedAppChanges(a.nsName, a.name, meta.LabelValue, a.appChangesUseAppLabel, a.coreClient).DeleteAll()
	if err != nil {
		return fmt.Errorf("Deleting app changes: %w", err)
	}

	err = app.Delete()
	if err != nil {
		return err
	}

	name := a.name
	if a.isMigrated {
		name = a.fqName()
	}

	err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Deleting app: %w", err)
	}

	return nil
}

func (a *RecordedApp) Rename(newName string, newNamespace string) error {
	name := a.name
	if a.isMigrated {
		name = a.fqName()
	}

	app, found, err := a.find(name)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf("App '%s' (namespace: %s) does not exist: %s",
			a.name, a.nsName, a.appInDiffNsHintMsgFunc(name))
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
		return fmt.Errorf("Creating app: %w", err)
	}

	err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Delete(context.TODO(), oldName, metav1.DeleteOptions{})
	if err != nil {
		// TODO Do not clean up new config map as there is no gurantee it can be deleted either
		return fmt.Errorf("Deleting app: %w", err)
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

func (a *RecordedApp) isMigrationEnabled() bool {
	return strings.ToLower(os.Getenv("KAPP_FQ_CONFIGMAP_NAMES")) == "true"
}

func (a *RecordedApp) Meta() (Meta, error) { return a.meta() }

func (a *RecordedApp) setMeta(app corev1.ConfigMap) (Meta, error) {
	// TODO: Should this be moved out of this function?
	_, a.appChangesUseAppLabel = app.Annotations[KappAppChangesUseAppLabelAnnotationKey]

	meta, err := NewAppMetaFromData(app.Data)
	if err != nil {
		errMsg := "App '%s' (namespace: %s) backed by ConfigMap '%s' did not contain parseable app metadata: %w"
		hintText := " (hint: ConfigMap was overriden by another user?)"

		if a.isMigrated {
			return Meta{}, fmt.Errorf(errMsg+hintText, a.name, a.nsName, a.fqName(), err)
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

	app, foundMigratedApp, err := a.find(a.fqName())
	if err != nil {
		return Meta{}, err
	}

	if foundMigratedApp {
		a.isMigrated = true
		return a.setMeta(*app)
	}

	app, foundNonMigratedApp, err := a.find(a.name)
	if err != nil {
		return Meta{}, err
	}

	if !foundNonMigratedApp {
		return Meta{}, fmt.Errorf("App '%s' (namespace: %s) does not exist: %s",
			a.name, a.nsName, a.appInDiffNsHintMsgFunc(a.name))
	}

	return a.setMeta(*app)
}

func (a *RecordedApp) Changes() ([]Change, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	return NewRecordedAppChanges(a.nsName, a.name, meta.LabelValue, a.appChangesUseAppLabel, a.coreClient).List()
}

func (a *RecordedApp) LastChange() (Change, error) {
	meta, err := a.meta()
	if err != nil {
		return nil, err
	}

	if meta.LastChange.Successful == nil {
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

func (a *RecordedApp) BeginChange(meta ChangeMeta, appChangesMaxToKeep int) (Change, error) {
	appMeta, err := a.meta()
	if err != nil {
		return nil, err
	}

	change, err := NewRecordedAppChanges(a.nsName, a.name, appMeta.LabelValue, a.appChangesUseAppLabel, a.coreClient).Begin(meta, appChangesMaxToKeep)
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
		name = a.fqName()
	}

	change, err := a.coreClient.CoreV1().ConfigMaps(a.nsName).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Getting app: %w", err)
	}

	meta, err := NewAppMetaFromData(change.Data)
	if err != nil {
		return err
	}

	doFunc(&meta)

	change.Data = meta.AsData()

	_, err = a.setMeta(*change)
	if err != nil {
		return err
	}

	_, err = a.coreClient.CoreV1().ConfigMaps(a.nsName).Update(context.TODO(), change, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Updating app: %w", err)
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
