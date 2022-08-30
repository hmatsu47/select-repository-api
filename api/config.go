package api

import (
	"bufio"
	"fmt"
	"os"
)

type Config struct {
	ServiceName			[]string
	RepositoryName		map[string][]string
	ServiceSettingPath	string
}

// サービスのリポジトリ一覧取得
func RepositoryNameList(settingPath string, serviceName string) []string {
	repositoriesFile := fmt.Sprintf("%s/%s-repositories", settingPath, serviceName)
	f, err := os.Open(repositoriesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "サービス%sのリポジトリ一覧ファイル（%s）の読み取りに失敗しました\n: %s", serviceName, repositoriesFile, err)
		os.Exit(1)
	}
	defer f.Close()

	rscanner := bufio.NewScanner(f)
	var repositoryNameList []string
	for rscanner.Scan() {
		rname := rscanner.Text()
		fmt.Printf("サービス（%s）のリポジトリ追加 : %s\n", serviceName, rname)
		repositoryNameList = append(repositoryNameList, rname)
	}
	return repositoryNameList
}

func NewConfig(workDir string) Config {
	// サービス設定パスは Working Directory とする（指定がない場合は /var/select-repository）
	settingPath := workDir
	if settingPath == "" {
		settingPath = "/var/select-repository"
	}
	// サービス設定パスにある services ファイルからサービス一覧を取得
	servicesFile := fmt.Sprintf("%s/services", settingPath)
	f, err := os.Open(servicesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "サービス一覧ファイル（%s）の読み取りに失敗しました\n: %s", servicesFile, err)
		os.Exit(1)
	}
	defer f.Close()

	repositoryNameMap := map[string][]string{}

	scanner := bufio.NewScanner(f)
	var serviceNameList []string
	for scanner.Scan() {
		name := scanner.Text()
		fmt.Printf("サービス追加 : %s\n", name)
		serviceNameList = append(serviceNameList, name)
		// 各サービスのリポジトリ一覧を取得
		repositoryNameMap[name] = RepositoryNameList(settingPath, name)
	}

	return Config {
		ServiceName: serviceNameList,
		RepositoryName: repositoryNameMap,
		ServiceSettingPath: settingPath,
	}
}
