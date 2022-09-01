package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ServiceName        []string
	RepositoryMap      map[string][]Repository
	ServiceSettingPath string
}

// サービスのリポジトリ一覧取得
func RepositoryList(settingPath string, serviceName string) []Repository {
	repositoriesFile := fmt.Sprintf("%s/%s-repositories", settingPath, serviceName)
	f, err := os.Open(repositoriesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "サービス%sのリポジトリ一覧ファイル（%s）の読み取りに失敗しました\n: %s", serviceName, repositoriesFile, err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var repositoryList []Repository
	for scanner.Scan() {
		uri := scanner.Text()
		name := strings.Split(uri, "/")[1]
		repo := Repository{
			Name: name,
			Uri:  uri,
		}
		repositoryList = append(repositoryList, repo)
		fmt.Printf("サービス（%s）のリポジトリ追加 : %s\n", serviceName, uri)
	}
	return repositoryList
}

func ReadConfig(workDir string) *Config {
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

	// サービス別リポジトリマップを取得
	repositoryMap := map[string][]Repository{}

	scanner := bufio.NewScanner(f)
	var serviceNameList []string
	for scanner.Scan() {
		name := scanner.Text()
		serviceNameList = append(serviceNameList, name)
		fmt.Printf("サービス追加 : %s\n", name)
		// 各サービスのリポジトリ一覧を取得
		repositoryMap[name] = RepositoryList(settingPath, name)
	}

	// サービス別リポジトリマップから全サービスリポジトリマップをせいせい

	return &Config{
		ServiceName:        serviceNameList,
		RepositoryMap:      repositoryMap,
		ServiceSettingPath: settingPath,
	}
}
