package mysql

import (
	"context"
	"database/sql"
	"dataflow/pkg/domain/repository"
	"dataflow/pkg/terrors"

	"github.com/volatiletech/sqlboiler/boil"
	"google.golang.org/grpc/codes"
)

type dbMasterTxManager struct {
	db *sql.DB
}

func NewDBMasterTxManager(db *sql.DB) repository.MasterTxManager {
	return &dbMasterTxManager{db}
}

func (m *dbMasterTxManager) Transaction(ctx context.Context, f func(ctx context.Context, masterTx repository.MasterTx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return terrors.Stack(err)
	}
	defer func() {
		// panic
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				err = terrors.Wrapf(e, codes.Internal, "Mysqlのトランザクションロールバックに失敗しました。", "failed to MySQL Rollback")
			}
			panic(p) // re-throw panic after Rollback
		}
		// error
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				err = terrors.Wrapf(e, codes.Internal, "Mysqlのトランザクションロールバックに失敗しました。", "failed to MySQL Rollback")
			}
		}
		// 正常
		e := tx.Commit()
		if e != nil {
			err = terrors.Wrapf(e, codes.Internal, "Mysqlのトランザクションコミットに失敗しました。", "failed to MySQL Commit")
		}
	}()
	err = f(ctx, &dbMasterTx{tx})
	if err != nil {
		return terrors.Stack(err)
	}
	return nil
}

type dbMasterTx struct {
	tx *sql.Tx
}

func (m *dbMasterTx) Commit() error {
	return m.tx.Commit()
}

func (m *dbMasterTx) Rollback() error {
	return m.tx.Rollback()
}

func ExtractExecutor(masterTx repository.MasterTx) (boil.ContextExecutor, error) {
	return ExtractTx(masterTx)
}

func ExtractTx(masterTx repository.MasterTx) (*sql.Tx, error) {
	// キャストする
	tx, ok := masterTx.(*dbMasterTx)
	if !ok {
		return nil, terrors.Newf(codes.Internal, "masterTxからdbMasterTxへのキャストに失敗しました。", "masterTx cannot cast to dbMasterTx")
	}
	return tx.tx, nil
}
