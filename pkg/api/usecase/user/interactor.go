package user

import (
	"context"
	"dataflow/pkg/domain/entity"
	"dataflow/pkg/domain/repository"
	userservice "dataflow/pkg/domain/service/user"
	"dataflow/pkg/terrors"
)

type Interactor interface {
	CreateNewUser(ctx context.Context, uid, name, thumbnail string) (*entity.User, error)
	GetUserProfile(ctx context.Context, uid string) (*entity.User, error)
	GetAll(ctx context.Context) (entity.UserSlice, error)
}

type intereractor struct {
	masterTxManager repository.MasterTxManager
	userService     userservice.Service
}

func New(masterTxManager repository.MasterTxManager, userService userservice.Service) Interactor {
	return &intereractor{
		masterTxManager: masterTxManager,
		userService:     userService,
	}
}

func (i *intereractor) CreateNewUser(ctx context.Context, uid, name, thumbnail string) (*entity.User, error) {
	var userData *entity.User
	var err error

	err = i.masterTxManager.Transaction(ctx, func(ctx context.Context, masterTx repository.MasterTx) error {
		// 新規ユーザ作成
		userData, err = i.userService.CreateNewUser(ctx, masterTx, uid, name, thumbnail)
		if err != nil {
			return terrors.Stack(err)
		}
		return nil
	})
	if err != nil {
		return nil, terrors.Stack(err)
	}
	return userData, nil
}

func (i *intereractor) GetUserProfile(ctx context.Context, uid string) (*entity.User, error) {
	var userData *entity.User
	var err error

	err = i.masterTxManager.Transaction(ctx, func(ctx context.Context, masterTx repository.MasterTx) error {
		// ログイン済ユーザのプロフィール情報取得
		userData, err = i.userService.GetByUID(ctx, masterTx, uid)
		if err != nil {
			return terrors.Stack(err)
		}
		return nil
	})
	if err != nil {
		return nil, terrors.Stack(err)
	}
	return userData, nil
}

func (i *intereractor) GetAll(ctx context.Context) (entity.UserSlice, error) {
	var userSlice entity.UserSlice
	var err error

	err = i.masterTxManager.Transaction(ctx, func(ctx context.Context, masterTx repository.MasterTx) error {
		// (管理者用)ユーザ全件取得
		userSlice, err = i.userService.GetAll(ctx, masterTx)
		if err != nil {
			return terrors.Stack(err)
		}
		return nil
	})
	if err != nil {
		return nil, terrors.Stack(err)
	}
	return userSlice, nil
}
