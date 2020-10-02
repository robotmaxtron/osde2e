package verify

import (
	"context"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	userv1 "github.com/openshift/api/user/v1"
	autoscalingv1 "github.com/openshift/cluster-autoscaler-operator/pkg/apis/autoscaling/v1"
	"github.com/openshift/osde2e/pkg/common/alert"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/helper"
	"github.com/openshift/osde2e/pkg/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	NvidiaGPU          = "nvidia.com/gpu"
	TestNamespace      = "test-namespace"
	TestCloudProvider  = "testProvider"
	TestReleaseVersion = "v100"
)

var (
	ScaleDownUnneededTime        = "10s"
	ScaleDownDelayAfterAdd       = "60s"
	MaxNodeProvisionTime         = "30m"
	PodPriorityThreshold   int32 = -10
	MaxPodGracePeriod      int32 = 60
	MaxNodesTotal          int32 = 100
	CoresMin               int32 = 16
	CoresMax               int32 = 32
	MemoryMin              int32 = 32
	MemoryMax              int32 = 64
	NvidiaGPUMin           int32 = 4
	NvidiaGPUMax           int32 = 8
)

var userWebhookTestName string = "[Suite: service-definition] [OSD] regularuser validating webhook"

func init() {
	alert.RegisterGinkgoAlert(userWebhookTestName, "SD-SREP", "Max Whittingham", "sd-cicd-alerts", "sd-cicd@redhat.com", 4)
}

var _ = ginkgo.Describe(userWebhookTestName, func() {
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
			_, err := NewClusterAutoscaler()(*autoscalingv1.ClusterAutoscaler, err)
			Expect(err).NotTo(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))

		ginkgo.It("unpriv users cannot create autoscalers", func() {
			userName := util.RandomStr(5) + "@customdomain"
			user, err := createUser(userName, []string{}, h)
			defer func() {
				h.Impersonate(rest.ImpersonationConfig{})
				h.User().UserV1().Users().Delete(context.TODO(), user.Name, metav1.DeleteOptions{})
			}()

			h.Impersonate(rest.ImpersonationConfig{
				UserName: "test@customdomain",
				Groups: []string{
					"system:authenticated",
					"system:authenticated:oauth",
				},
			})
			_, err = NewClusterAutoscaler()(*autoscalingv1.ClusterAutoscaler, err)
			Expect(err).To(HaveOccurred())
		}, float64(viper.GetFloat64(config.Tests.PollingTimeout)))
	})
})

func createUser(userName string, groups []string, h *helper.H) (*userv1.User, error) {
	user := &userv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Groups: groups,
	}
	return h.User().UserV1().Users().Create(context.TODO(), user, metav1.CreateOptions{})
}

//Pulled from cluster-autoscaler-operator
func NewClusterAutoscaler() *autoscalingv1.ClusterAutoscaler {
	return &autoscalingv1.ClusterAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAutoscaler",
			APIVersion: "autoscaling.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: TestNamespace,
		},
		Spec: autoscalingv1.ClusterAutoscalerSpec{
			MaxPodGracePeriod:    &MaxPodGracePeriod,
			PodPriorityThreshold: &PodPriorityThreshold,
			ResourceLimits: &autoscalingv1.ResourceLimits{
				MaxNodesTotal: &MaxNodesTotal,
				Cores: &autoscalingv1.ResourceRange{
					Min: CoresMin,
					Max: CoresMax,
				},
				Memory: &autoscalingv1.ResourceRange{
					Min: MemoryMin,
					Max: MemoryMax,
				},
				GPUS: []autoscalingv1.GPULimit{
					{
						Type: NvidiaGPU,
						Min:  NvidiaGPUMin,
						Max:  NvidiaGPUMax,
					},
				},
			},
			ScaleDown: &autoscalingv1.ScaleDownConfig{
				Enabled:       true,
				DelayAfterAdd: &ScaleDownDelayAfterAdd,
				UnneededTime:  &ScaleDownUnneededTime,
			},
		},
	}
}
