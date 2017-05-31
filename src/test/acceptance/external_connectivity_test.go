package acceptance_test

import (
	"cf-pusher/cf_cli_adapter"
	"fmt"
	"io/ioutil"
	"lib/testsupport"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("external connectivity", func() {
	var (
		appA                          string
		orgName                       string
		spaceName                     string
		appRoute                      string
		prefix                        string
		originalRunningSecurityGroups []string
		cli                           *cf_cli_adapter.Adapter
	)

	BeforeEach(func() {
		cli = &cf_cli_adapter.Adapter{CfCliPath: "cf"}
		appA = fmt.Sprintf("appA-%d", rand.Int31())
		prefix = testConfig.Prefix

		AuthAsAdmin()

		orgName = prefix + "external-connectivity-org"
		Expect(cf.Cf("create-org", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))

		spaceName = prefix + "space"
		Expect(cf.Cf("create-space", spaceName, "-o", orgName).Wait(Timeout_Push)).To(gexec.Exit(0))
		Expect(cf.Cf("target", "-o", orgName, "-s", spaceName).Wait(Timeout_Push)).To(gexec.Exit(0))

		By("discovering all existing running ASGs")
		originalRunningSecurityGroups = getRunningSecurityGroups()

		By("unbinding all running ASGs")
		for _, sg := range originalRunningSecurityGroups {
			Expect(cf.Cf("unbind-running-security-group", sg).Wait(Timeout_Short)).To(gexec.Exit(0))
		}

		By("creating test-generated ASGs")
		for asgName, asgValue := range testASGs {
			createASG(asgName, asgValue)
		}

		By("pushing the test app")
		pushProxy(appA)
		appRoute = fmt.Sprintf("http://%s.%s/", appA, config.AppsDomain)
	})

	AfterEach(func() {
		appReport(appA, Timeout_Short)
		Expect(cf.Cf("delete-org", orgName, "-f").Wait(Timeout_Push)).To(gexec.Exit(0))

		By("adding back all the original running ASGs")
		for _, sg := range originalRunningSecurityGroups {
			Expect(cf.Cf("bind-running-security-group", sg).Wait(Timeout_Short)).To(gexec.Exit(0))
		}

		By("removing test-generated ASGs")
		for asgName, _ := range testASGs {
			removeASG(asgName)
		}
	})

	checkRequest := func(route string, expectedStatusCode int, expectedResponseSubstring string) error {
		resp, err := http.Get(route)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		respBody := string(respBytes)

		if resp.StatusCode != expectedStatusCode {
			return fmt.Errorf("test http get to %s: expected response code %d but got %d.  response body:\n%s", route, expectedStatusCode, resp.StatusCode, respBody)
		}
		if !strings.Contains(respBody, expectedResponseSubstring) {
			return fmt.Errorf("test http get to %s: expected response to contain %q but instead saw:\n%s", route, expectedResponseSubstring, respBody)
		}
		return nil
	}

	canProxy := func() error {
		return checkRequest(appRoute+"proxy/example.com", 200, "Example Domain")
	}
	isReachable := func() error {
		return checkRequest(appRoute, 200, `{"ListenAddresses":[`)
	}
	canPing := func() error {
		return checkRequest(appRoute+"ping/example.com", 200, "Ping succeeded")
	}
	cannotProxy := func() error {
		return checkRequest(appRoute+"proxy/example.com", 500, "example.com")
	}
	cannotPing := func() error {
		return checkRequest(appRoute+"ping/example.com", 500, "Ping failed to destination: example.com")
	}

	Describe("basic (legacy) network behavior for an app", func() {
		It("makes the app reachable from the router, and the app can reach the internet only if allowed", func(done Done) {
			By("checking that the app is reachable via the router")
			Eventually(isReachable, "10s", "1s").Should(Succeed())
			Consistently(isReachable, "2s", "0.5s").Should(Succeed())

			By("checking that the app cannot reach the internet using http and dns")
			Eventually(cannotProxy, "10s", "1s").Should(Succeed())
			Consistently(cannotProxy, "2s", "0.5s").Should(Succeed())

			By("checking that the app cannot ping the internet (first time)")
			Consistently(cannotPing, "2s", "0.5s").Should(Succeed())

			By("creating and binding a tcp and udp security group")
			Expect(cli.BindSecurityGroup("tcp-asg", orgName, spaceName)).To(Succeed())
			Expect(cli.BindSecurityGroup("udp-asg", orgName, spaceName)).To(Succeed())

			By("restarting the app")
			Expect(cf.Cf("restart", appA).Wait(Timeout_Push)).To(gexec.Exit(0))

			By("checking that the app can use dns and http to reach the internet")
			Eventually(canProxy, "10s", "1s").Should(Succeed())
			Consistently(canProxy, "2s", "0.5s").Should(Succeed())

			By("checking that the app cannot ping the internet (second time)")
			Consistently(cannotPing, "2s", "1s").Should(Succeed())

			By("creating and binding an icmp security group")
			Expect(cli.BindSecurityGroup("icmp-asg", orgName, spaceName)).To(Succeed())

			By("removing the tcp security groups")
			Expect(cli.UnbindSecurityGroup("tcp-asg", orgName, spaceName)).To(Succeed())

			By("restarting the app")
			Expect(cf.Cf("restart", appA).Wait(Timeout_Push)).To(gexec.Exit(0))

			By("checking that the app can ping the internet")
			Eventually(canPing, "10s", "1s").Should(Succeed())
			Consistently(canPing, "2s", "0.5s").Should(Succeed())

			By("checking that the app cannot use http to reach the internet")
			Consistently(cannotProxy, "2s", "0.5s").Should(Succeed())

			close(done)
		}, 180 /* <-- overall spec timeout in seconds */)
	})
})

func getRunningSecurityGroups() []string {
	session := cf.Cf("running-security-groups")
	Expect(session.Wait(Timeout_Short)).To(gexec.Exit(0))

	candidateGroups := strings.Split(string(session.Out.Contents()), "\n")[3:]
	actualGroups := []string{}
	for _, l := range candidateGroups {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			actualGroups = append(actualGroups, trimmed)
		}
	}
	return actualGroups
}

func getAllSecurityGroups() []string {
	session := cf.Cf("security-groups")
	Expect(session.Wait(Timeout_Short)).To(gexec.Exit(0))

	candidateGroups := strings.Split(string(session.Out.Contents()), "\n")[4:]
	actualGroups := []string{}
	for _, l := range candidateGroups {
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		trimmed := strings.TrimSpace(fields[1])
		if trimmed != "" {
			actualGroups = append(actualGroups, trimmed)
		}
	}
	return actualGroups
}

func createASG(name string, asgDefinition string) {
	asgFile, err := testsupport.CreateASGFile(asgDefinition)
	Expect(err).NotTo(HaveOccurred())
	Expect(cli.CreateSecurityGroup(name+"-asg", asgFile)).To(Succeed())
	Expect(os.Remove(asgFile)).To(Succeed())
}

func removeASG(name string) {
	Expect(cli.DeleteSecurityGroup(name + "-asg")).To(Succeed())
}

var testASGs = map[string]string{
	"tcp": `
	[
		{
			"destination": "0.0.0.0-255.255.255.255",
			"protocol": "tcp",
			"ports": "80,443"
		}
	]
	`,
	"udp": `
	[
		{
			"destination": "0.0.0.0-255.255.255.255",
			"protocol": "udp",
			"ports": "53"
		}
	]
	`,
	"icmp": `
	[
		{
			"destination": "0.0.0.0-255.255.255.255",
			"protocol": "icmp",
			"type": 8,
			"code": 0
		}
	]
	`,
}
