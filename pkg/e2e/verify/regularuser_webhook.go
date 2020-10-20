package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

var regularuserWebhookTestName string = "[Suite: service-definition] [OSD] regularuser validating webhook"

func init() {
	alert.RegisterGinkgoAlert(regularuserWebhookTestName, "SD-SREP", "Max Whittingham", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(regularuserWebhookTestName, func() {
	h := helper.New()

	ginkgo.Context("regularuser validating webhook", func() {
		ginkgo.It("kube:system can create autoscalers", func() {
			h.Impersonate(rest.ImpersonationConfig{
				UserName: "kube:system",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			_, err := h.Dynamic().Resource(schema.GroupVersionResource{
				Group:    "autoscaling.openshift.io",
				Version:  "v1",
				Resource: "ClusterAutoscaler"}).Create(context.TODO(), &unstructured.Unstructured{}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("unpriv users cannot create autoscalers", func() {
			userName := util.RandomStr(5) + "@customdomain"
			userRandom, err := createUser(userName, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				h.User().UserV1().Users().Delete(context.TODO(), userRandom.Name, metav1.DeleteOptions{})
			}()

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			_, err = h.Dynamic().Resource(schema.GroupVersionResource{
				Group:    "autoscaling.openshift.io",
				Version:  "v1",
				Resource: "ClusterAutoscaler"}).Create(context.TODO(), &unstructured.Unstructured{}, metav1.CreateOptions{})
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})
