package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

var (
	namespace string
	help      bool
)

func init() {
	pflag.StringVarP(&namespace, "namespace", "n", "", "indicate namespace")
	pflag.BoolVarP(&help, "help", "h", false, "help info")
	pflag.Usage = func() {
		fmt.Println(`Usage: bp [-n NAMESPACE ] 
Describe: 
  Expand ClusterRoleBinding or RoleBinding detail info!
Example:
  expand clusterrolebinding if If you don't specify -n params:
  ]# bp
  expand rolebinding if If you specify -n params (default: ""):
  ]# bp -n NAMESPACE
`)
	}
	pflag.Parse()
	if help {
		pflag.Usage()
		os.Exit(1)
	}
}
func getClusterRoleBindingOrRoleBinding() (interface{}, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	kubeConfigPath := filepath.Join(dir, ".kube", "config")
	if !fileExists(kubeConfigPath) {
		panic(errors.New("kubeConfig file not exists"))
	}
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	//test local
	//restConfig, err := clientcmd.BuildConfigFromFlags("", "../config")
	if err != nil {
		panic(err)
	}
	restClient, err := kubernetes.NewForConfig(restConfig)

	if strings.TrimSpace(namespace) == "" {
		return restClient.RbacV1().ClusterRoleBindings().List(context.Background(), metav1.ListOptions{})
	}
	return restClient.RbacV1().RoleBindings(namespace).List(context.Background(), metav1.ListOptions{})
}

func handleData() (data [][]string) {
	items, err := getClusterRoleBindingOrRoleBinding()
	if err != nil {
		panic(err)
	}
	data = make([][]string, 0)
	switch items.(type) {
	case *v1.ClusterRoleBindingList:
		for _, item := range items.(*v1.ClusterRoleBindingList).Items {
			column := make([]string, 0)
			if len(item.Subjects) == 0 {
				item.Subjects = append(item.Subjects, v1.Subject{
					Name: "none",
					Kind: "none",
				})
			}
			column = append(column, item.Name, item.RoleRef.Name, item.Subjects[0].Kind, item.Subjects[0].Name)
			data = append(data, column)
		}
	case *v1.RoleBindingList:
		for _, item := range items.(*v1.RoleBindingList).Items {
			column := make([]string, 0)
			if len(item.Subjects) == 0 {
				item.Subjects = append(item.Subjects, v1.Subject{
					Name: "none",
					Kind: "none",
				})
			}
			column = append(column, item.Name, item.RoleRef.Name, item.Subjects[0].Kind, item.Subjects[0].Name)
			data = append(data, column)
		}
	}
	return data
}

func main() {
	// Change table lines
	printColumns(handleData())
}

func printColumns(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetNoWhiteSpace(true)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetRowLine(false)
	table.SetTablePadding("\t")
	table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"BindingName", "RoleName", "SujKind", "SujName"})
	table.AppendBulk(data)
	table.Render()
}

func fileExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if os.IsNotExist(err) {
		return false
	}
	return true
}
