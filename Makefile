SOURCE_FILE := $(notdir $(source))
SOURCE_DIR := $(dir $(source))
MOCK_FILE := mock_${SOURCE_FILE}
MOCK_DIR := ${SOURCE_DIR}mock_$(lastword $(subst /, ,${SOURCE_DIR}))/
MOCK_TARGET := $(lastword $(subst /, ,${SOURCE_DIR}))
PROTOS_DIR := ./protobuf

GOLINT_FILE_NAME := golangci-lint
GOLINT_FILE_PATH := $(shell ls $(GOPATH)/bin | grep ${GOLINT_FILE_NAME})
SQLBOILER_FILE_NAME := sqlboiler
SQLBOILER_FILE_PATH := $(shell ls $(GOPATH)/bin | grep ${SQLBOILER_FILE_NAME})

# gocli-lintパッケージがダウンロード済みであればダウンロードを実行しない
define golintExist
    $(ifneq (${GOLINT_FILE_PATH},${GOLINT_FILE_NAME}),GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint)
endef

help: ## 使い方
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

protoc: ## protoファイルから自動生成
	protoc \
            -I ${PROTOS_DIR} \
            --go_out=plugins=grpc:pkg/pb/ \
            ${PROTOS_DIR}/*.proto && \
    protoc \
    		-I ${PROTOS_DIR} \
    		--doc_out=html,index.html:./proto-document \
    		 ${PROTOS_DIR}/*.proto \

mockgen: # mockgenの実行
	# Usege: make mockgen source=<インターフェースの定義しているファイル>

	# mockgenのインストール
	go install github.com/golang/mock/mockgen

	# mockgenの実行
	mockgen -source ${SOURCE_DIR}${SOURCE_FILE} -destination ${MOCK_DIR}${MOCK_FILE}

dbgen: ## sqlboilerによるコード自動生成
	# sqlboilerのインストール
	go get -u -t github.com/volatiletech/sqlboiler && \
 	go get github.com/volatiletech/sqlboiler/drivers/sqlboiler-mysql

	# DDL定義を元にコードを自動生成
	sqlboiler mysql -o db/mysql/model -p model --wipe db/mysql/model

wiregen: ## wire_gen.goの生成
	# google/wireのインストール
	GO111MODULE=off go get -u github.com/google/wire

	# wire genの実行
	wire gen cmd/wire.go

test: ## testの実行
	go test -v ./...

lint: ## lintの実行
	# golangci-lintのインストール(既にパッケージがあれば実行されない)
	${golintExist}

	# pkg配下をチェック。設定は .golangci.yml に記載
	golangci-lint run

fmt: ## fmtの実行
	# goimportsのインストール
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports

	# tidy, fmt, goimportsの実行
	go mod tidy -v
	gofmt -s -w pkg/
	goimports -w pkg/

fmt-lint: fmt lint ## fmtとlintの実行

run: ## APIをビルドせずに立ち上げるコマンド
	go run ./cmd

build: ## APIをビルドして立ち上げるコマンド
	go build -o binary/dataflow ./cmd
	./binary/dataflow