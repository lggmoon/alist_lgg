package db

import (
	"fmt"
	"sort"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

// DeleteStorageById just delete storage from database by id
func DeleteStorageById_user(user *model.User, id uint) error {
	return errors.WithStack(db.Where("user_id=?", user.ID).Delete(&model.Storage{}, id).Error)
}

// GetStorages Get all storages from database order by index
func GetStorages_user(user *model.User, pageIndex, pageSize int) ([]model.Storage, int64, error) {
	storageDB := db.Model(&model.Storage{})
	var count int64
	if err := storageDB.Where("user_id=?", user.ID).Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get storages count")
	}
	var storages []model.Storage
	if err := storageDB.Where("user_id=?", user.ID).Order(columnName("order")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&storages).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return storages, count, nil
}

// GetStorageById Get Storage by id, used to update storage usually
func GetStorageById_user(user *model.User, id uint) (*model.Storage, error) {
	var storage model.Storage
	storage.ID = id
	storage.UserID = user.ID
	if err := db.First(&storage).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &storage, nil
}

// GetStorageByMountPath Get Storage by mountPath, used to update storage usually
func GetStorageByMountPath_user(user *model.User, mountPath string) (*model.Storage, error) {
	var storage model.Storage
	if err := db.Where("user_id=?", user.ID).Where("mount_path = ?", mountPath).First(&storage).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &storage, nil
}

func GetEnabledStorages_user(user *model.User) ([]model.Storage, error) {
	var storages []model.Storage
	if err := db.Where("user_id=?", user.ID).Where(fmt.Sprintf("%s = ?", columnName("disabled")), false).Find(&storages).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	sort.Slice(storages, func(i, j int) bool {
		return storages[i].Order < storages[j].Order
	})
	return storages, nil
}
