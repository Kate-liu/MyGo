package main

import (
	// go 标准包
	"fmt"

	// 第三方包
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// 匿名包单独分组，并对匿名包引用进行说明
	// import mysql driver
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// 内部包
	v1 "github.com/marmotedu/api/apiserver/v1"
	metav1 "github.com/marmotedu/apimachinery/pkg/meta/v1"
	"github.com/marmotedu/iam/pkg/cli/genericclioptions"
)
