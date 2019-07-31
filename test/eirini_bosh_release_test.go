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
				session = boshRunErrand("configure-eirini-bosh", "5m", expectedExitCode)
			})

			Context("configured to reference an image that does not exist", func() {
				BeforeEach(func() {
					boshDeployOpsFiles = []string{"operations/invalid-image-reference-for-errand.yml"}
					expectedExitCode = 1
				})

				It("the errand should error out with a meaningful message", func() {
					Expect(session).Should(Say("Error: ImagePullBackOff"))
				})
			})

			Context("configured with an invalid service account", func() {
				BeforeEach(func() {
					boshDeployOpsFiles = []string{"operations/invalid-service-account-for-errand.yml"}
					expectedExitCode = 1
				})

				It("the errand should error out with a meaningful message", func() {
					Expect(session).Should(Say("Unauthorized"))
				})
			})

			FContext("configured correctly", func() {
				Context("and the errand has been running longer than a configurable timeout", func() {
					BeforeEach(func() {
						boshDeployOpsFiles = []string{"operations/short-timeout-for-errand.yml"}
						expectedExitCode = 1
					})

					It("the errand should error out because of the configured timeout", func() {
						Expect(session).Should(Say("Unauthorized"))
					})
				})
			})
		})
	})
})
