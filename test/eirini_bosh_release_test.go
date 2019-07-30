package bosh_release_tests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("EiriniBoshRelease", func() {
	var (
		boshDeployOpsFiles []string

		session *gexec.Session

		expectedExitCode int
		errandTimeout    string
	)

	Describe("when eirini has been BOSH-deployed successfully", func() {
		JustBeforeEach(func() {
			boshDeploy(boshDeployOpsFiles...)
		})

		// AfterEach(func() {
		// 	boshDeleteDeployment()
		// })

		Describe("and I run the configure-eirini-bosh errand", func() {
			JustBeforeEach(func() {
				session = boshRunErrand("configure-eirini-bosh", errandTimeout, expectedExitCode)
			})

			Context("configured to reference an image that does not exist", func() {
				BeforeEach(func() {
					boshDeployOpsFiles = []string{"operations/invalid-image-reference-for-errand.yml"}
					errandTimeout = "5m"
					expectedExitCode = 1
				})

				It("the errand should error out with a meaningful message", func() {
					Expect(session).Should(Say("Error: ImagePullBackOff"))
				})
			})

			Context("configured with an invalid service account", func() {})

			Context("configured correctly", func() {
				Context("and the errand has been running longer than a configured timeout", func() {})
			})
		})
	})
})
