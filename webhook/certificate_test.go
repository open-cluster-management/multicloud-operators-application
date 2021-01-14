// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8scertutil "k8s.io/client-go/util/cert"
)

var _ = Describe("self-sign cert", func() {
	It("should generate CA cert and persist to secret", func() {
		podNs := "test"

		os.Setenv(podNamespaceEnvVar, podNs)

		certDir := "/tmp/tmp-cert"

		defer func() {
			os.RemoveAll(certDir)
			os.Unsetenv(podNamespaceEnvVar)
		}()

		ca, err := GenerateWebhookCerts(k8sClient, certDir)
		Expect(err).Should(Succeed())
		Expect(ca).ShouldNot(BeNil())

		canReadCertAndKey, err := k8scertutil.CanReadCertAndKey("/tmp/tmp-cert/tls.crt", "/tmp/tmp-cert/tls.key")
		Expect(err).Should(Succeed())
		Expect(canReadCertAndKey).Should(BeTrue())

		srtKey := types.NamespacedName{Name: WebhookServiceName, Namespace: podNs}

		srtIns := &corev1.Secret{}
		Expect(k8sClient.Get(context.TODO(), srtKey, srtIns)).Should(Succeed())
		defer func() {
			Expect(k8sClient.Delete(context.TODO(), srtIns)).Should(Succeed())
		}()

		Expect(srtIns.Data["crt"]).ShouldNot(HaveLen(0))
		Expect(srtIns.Data["key"]).ShouldNot(HaveLen(0))
	})

	It("should get self-signed CA cert from exist secret", func() {
		podNs := "test"

		os.Setenv(podNamespaceEnvVar, podNs)
		defer func() {
			os.Unsetenv(podNamespaceEnvVar)
		}()

		cert := "my cert"
		key := "my key"
		srtIns := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      WebhookServiceName,
				Namespace: podNs,
			},

			Data: map[string][]byte{
				"crt": []byte(cert),
				"key": []byte(key),
			},
		}

		Expect(k8sClient.Create(context.TODO(), srtIns)).Should(Succeed())

		ca, err := getSelfSignedCACert(k8sClient, certName, types.NamespacedName{Name: srtIns.Name, Namespace: srtIns.Namespace})
		Expect(err).Should(Succeed())

		Expect(ca.Cert).Should(Equal(cert))
		Expect(ca.Key).Should(Equal(key))
	})
})
