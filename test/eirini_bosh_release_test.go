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
				session = boshRunErrand("configure-eirini-bosh")
			})

			Context("configured to reference an image that does not exist", func() {
				BeforeEach(func() {
					// 1. configure the errand
					// 1. deploy eirini
					// 1. run the errand
					boshDeployOpsFiles = []string{"operations/invalid-image-reference-for-errand.yml"}
				})

				It("the errand should error out with a meaningful message", func() {
					Eventually(session, "5m").ShouldNot(gexec.Exit(0))
					Expect(session).Should(Say("repository does not exist"))
				})
			})

			Context("configured with an invalid service account", func() {})

			Context("configured correctly", func() {
				Context("and the errand has been running longer than a configured timeout", func() {})
			})
		})
	})
})
