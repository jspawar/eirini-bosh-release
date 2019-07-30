package bosh_release_tests_test

import (
	"fmt"
	"testing"
	"time"

	"os/exec"

	"os"

	"io/ioutil"

	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	boshDeployCmd           = []string{"-d", "eirini", "deploy", "deployment.yml", "-n", "--no-redact"}
	boshDeleteDeploymentCmd = []string{"-d", "eirini", "delete-deployment", "-n"}

	kubeConfig *rest.Config

	kubeHostUrl             string
	kubeHostCa              string
	kubeServiceAccountName  string
	kubeServiceAccountToken string
	kubeNamespace           string
)

func TestBoshReleaseTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BoshReleaseTests Suite")
}

// TODO: create and upload release?
var _ = SynchronizedBeforeSuite(func() []byte { return nil }, func([]byte) {
	kubeConfig = getKubeConfigFromEnv()
	kubeNamespace = createTestKubeNamespace()
	kubeServiceAccountName, kubeServiceAccountToken = createTestServiceAccount()
})

// var _ = SynchronizedAfterSuite(func() {
// 	deleteTestKubeNamespace()
//  TODO: delete cluster role binding bc it isnt' namespaced!!
// }, func() {})

func getKubeConfigFromEnv() *rest.Config {
	kubeConfigPath, varSet := os.LookupEnv("KUBECONFIG")
	Expect(varSet).To(BeTrue(), "KUBECONFIG must be set with current context using service account credentials")

	bs, err := ioutil.ReadFile(kubeConfigPath)
	Expect(err).To(BeNil())

	conf, err := clientcmd.RESTConfigFromKubeConfig(bs)
	Expect(err).To(BeNil())

	return conf
}

func createTestKubeNamespace() string {
	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	Expect(err).To(BeNil())

	testKubeNamespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("bosh-release-tests-%d-%d", time.Now().Unix(), GinkgoParallelNode())}}
	ns, err := kubeClientset.CoreV1().Namespaces().Create(testKubeNamespace)
	Expect(err).To(BeNil())

	return ns.Name
}

func createTestServiceAccount() (string, string) {
	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	Expect(err).To(BeNil())

	svcAccount, err := kubeClientset.CoreV1().ServiceAccounts(kubeNamespace).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "bosh-release-tests-service-account"},
	})
	Expect(err).To(BeNil())

	_, err = kubeClientset.RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("bosh-release-tests-service-account-%s-cluster-admin", kubeNamespace)},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      "bosh-release-tests-service-account",
			Namespace: kubeNamespace,
		}},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
	})
	Expect(err).To(BeNil())

	var secrets []v1.ObjectReference
	Eventually(func() []v1.ObjectReference {
		s, err := kubeClientset.CoreV1().ServiceAccounts(kubeNamespace).Get("bosh-release-tests-service-account", metav1.GetOptions{})
		Expect(err).To(BeNil())
		secrets = s.Secrets
		return secrets
	}).Should(HaveLen(1))
	svcAccountTokenSecretName := secrets[0].Name

	svcAccountTokenSecret, err := kubeClientset.CoreV1().Secrets(kubeNamespace).Get(svcAccountTokenSecretName, metav1.GetOptions{})
	Expect(err).To(BeNil())

	Expect(svcAccountTokenSecret.Data).To(HaveKey("token"))

	return svcAccount.Name, string(svcAccountTokenSecret.Data["token"])
}

func deleteTestKubeNamespace() {
	kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
	Expect(err).To(BeNil())

	Expect(kubeClientset.CoreV1().Namespaces().Delete(kubeNamespace, &metav1.DeleteOptions{})).To(Succeed())
}

// TODO: maybe rename this and delete the file after
func makestupiduglyvarsfile() string {
	garbage := map[string]string{
		"k8s_host_url":         kubeConfig.Host,
		"k8s_node_ca":          string(kubeConfig.TLSClientConfig.CAData),
		"k8s_system_namespace": kubeNamespace,
		"k8s_service_username": kubeServiceAccountName,
		"k8s_service_token":    kubeServiceAccountToken,
	}
	bs, err := json.Marshal(garbage)
	Expect(err).To(BeNil())

	f, err := ioutil.TempFile("", "bosh-deploy-vars-*.json")
	Expect(err).To(BeNil())

	Expect(ioutil.WriteFile(f.Name(), bs, 0644)).To(Succeed())

	return f.Name()
}

func boshDeploy(opsFiles ...string) {
	actualDeployCmd := make([]string, len(boshDeployCmd))
	copy(actualDeployCmd, boshDeployCmd)
	for _, opsFile := range opsFiles {
		actualDeployCmd = append(actualDeployCmd, "-o", opsFile)
	}
	actualDeployCmd = append(actualDeployCmd, "-l", makestupiduglyvarsfile())

	cmd := exec.Command("bosh", actualDeployCmd...)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).To(BeNil())
	Eventually(session, "5m").Should(gexec.Exit(0))
}

func boshRunErrand(errandName, timeout string, expectedStatusCode int) *gexec.Session {
	runErrandCmd := []string{"-d", "eirini", "run-errand", errandName}
	cmd := exec.Command("bosh", runErrandCmd...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).To(BeNil())
	Eventually(session, timeout).Should(gexec.Exit(expectedStatusCode))
	return session
}

func boshDeleteDeployment() {
	cmd := exec.Command("bosh", boshDeleteDeploymentCmd...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).To(BeNil())
	Eventually(session, "5m").Should(gexec.Exit(0))
}

