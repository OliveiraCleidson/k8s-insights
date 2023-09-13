package usecases

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

func GetDeploymentsConf() error {
	// var outputPath string

	deployment := exec.Command("kubectl", "get", "deployments", "--all-namespaces")

	output, err := deployment.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	deploymentsConf := make([]deploymentConf, 0)
	for i, line := range lines {
		if len(line) == 0 || i == 0 {
			continue
		}

		spaceRegex := regexp.MustCompile(`(\s+){1,}`)
		lineContent := spaceRegex.ReplaceAllString(line, ",")
		deployment := strings.Split(lineContent, ",")

		namespace := deployment[0]
		name := deployment[1]

		wg.Add(1)
		go func() {
			deploymentConf, err := getDeploymentConf(namespace, name)
			if err != nil {
				wg.Done()
				return
			}

			mutex.Lock()
			deploymentsConf = append(deploymentsConf, *deploymentConf)
			mutex.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	csvLines := make([]string, 0)
	csvLines = append(csvLines, "namespace,name,liveness,readiness,startup")

	for _, deploymentConf := range deploymentsConf {
		csvLines = append(csvLines, fmt.Sprintf("%s,%s,%s,%s,%s", deploymentConf.Namespace, deploymentConf.Name, deploymentConf.Liveness, deploymentConf.Readiness, deploymentConf.Startup))
	}

	file, err := os.Create("deployments.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range csvLines {
		file.WriteString(line + "\n")
	}

	return nil
}

type deploymentConf struct {
	Namespace string
	Name      string
	Liveness  string
	Readiness string
	Startup   string
}

func getDeploymentConf(namespace, deploymentName string) (*deploymentConf, error) {
	describe := exec.Command("kubectl", "describe", fmt.Sprintf("deployments/%s", deploymentName), "-n", namespace)

	out, err := describe.Output()
	if err != nil {
		return nil, err
	}

	livenessRegex := regexp.MustCompile(`Liveness:((\s){0,})(.+)`)
	readinessRegex := regexp.MustCompile(`Readiness:((\s){0,})(.+)`)
	startupRegex := regexp.MustCompile(`Startup:((\s){0,})(.+)`)

	livenessConf := "Sem configuração"
	readinessConf := "Sem configuração"
	startupConf := "Sem configuração"

	for _, line := range strings.Split(string(out), "\n") {
		if livenessRegex.MatchString(line) {
			livenessConf = livenessRegex.FindStringSubmatch(line)[3]
			continue
		}

		if readinessRegex.MatchString(line) {
			readinessConf = readinessRegex.FindStringSubmatch(line)[3]
			continue
		}

		if startupRegex.MatchString(line) {
			startupConf = startupRegex.FindStringSubmatch(line)[3]
		}
	}

	return &deploymentConf{
		Namespace: namespace,
		Name:      deploymentName,
		Liveness:  livenessConf,
		Readiness: readinessConf,
		Startup:   startupConf,
	}, nil

}
