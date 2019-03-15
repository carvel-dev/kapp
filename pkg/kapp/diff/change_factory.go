package diff

import (
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
)

type ChangeFactory struct {
	mods []ctlres.FieldCopyMod
}

func NewChangeFactory(mods []ctlres.FieldCopyMod) ChangeFactory {
	return ChangeFactory{mods}
}

func (f ChangeFactory) NewChangeAgainstLastApplied(existingRes, newRes ctlres.Resource) (Change, error) {
	if existingRes != nil {
		lastAppliedRes := f.NewResourceWithHistory(existingRes).LastAppliedResource()
		if lastAppliedRes != nil {
			rebasedLastAppliedRes, err := NewRebasedResource(existingRes, lastAppliedRes, f.mods).Resource()
			if err != nil {
				return nil, err
			}
			existingRes = rebasedLastAppliedRes
		}

		historylessExistingRes, err := f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	if newRes != nil {
		historylessNewRes, err := f.NewResourceWithHistory(newRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingRes, newRes, f.mods).Resource()
	if err != nil {
		return nil, err
	}

	return NewChange(existingRes, rebasedNewRes, newRes), nil
}

func (f ChangeFactory) NewExactChange(existingRes, newRes ctlres.Resource) (Change, error) {
	if existingRes != nil {
		historylessExistingRes, err := f.NewResourceWithHistory(existingRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		existingRes = historylessExistingRes
	}

	if newRes != nil {
		historylessNewRes, err := f.NewResourceWithHistory(newRes).HistorylessResource()
		if err != nil {
			return nil, err
		}

		newRes = historylessNewRes
	}

	rebasedNewRes, err := NewRebasedResource(existingRes, newRes, f.mods).Resource()
	if err != nil {
		return nil, err
	}

	return NewChange(existingRes, rebasedNewRes, newRes), nil
}

func (f ChangeFactory) NewResourceWithHistory(resource ctlres.Resource) ResourceWithHistory {
	return NewResourceWithHistory(resource, &f)
}
