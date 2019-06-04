// Copyright 2019 RedHat
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

package metrics

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Custom errors

// ErrMetricsFailedGenerateService indicates the metric service failed to generate
var ErrMetricsFailedGenerateService = errors.New("FailedGeneratingService")

// ErrMetricsFailedCreateService indicates that the service failed to create
var ErrMetricsFailedCreateService = errors.New("FailedCreateService")

// ErrMetricsFailedCreateServiceMonitor indicates that an account creation failed
var ErrMetricsFailedCreateServiceMonitor = errors.New("FailedCreateServiceMonitor")

// ErrMetricsFailedRegisterPromCRDs indicates that an account creation failed
var ErrMetricsFailedRegisterPromCRDs = errors.New("FailedCreateRegisterProm")

// GenerateService returns the static service which exposes specifed port.
func GenerateService(port int32, portName string) (*v1.Service, error) {
	operatorName, err := k8sutil.GetOperatorName()
	if err != nil {
		return nil, err
	}
	namespace, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorName,
			Namespace: namespace,
			Labels:    map[string]string{"name": operatorName},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port:     port,
					Protocol: v1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: port,
					},
					Name: portName,
				},
			},
			Selector: map[string]string{"name": operatorName},
		},
	}
	return service, nil
}

// GenerateServiceMonitor generates a prometheus-operator ServiceMonitor object
// based on the passed Service object.
func GenerateServiceMonitor(s *v1.Service) *monitoringv1.ServiceMonitor {
	labels := make(map[string]string)
	for k, v := range s.ObjectMeta.Labels {
		labels[k] = v
	}

	return &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceMonitor",
			APIVersion: "monitoring.coreos.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.ObjectMeta.Name,
			Namespace: s.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port: s.Spec.Ports[0].Name,
				},
			},
		},
	}
}

// ConfigureMetrics generates metrics service and servicemonitor,
// creates the metrics service and service monitor,
// and finally it starts the metrics server
func ConfigureMetrics(log logr.Logger, mgr manager.Manager) error {

	log.Info("Starting prometheus metrics")
	StartMetrics()

	if err := monitoringv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Info("Error registering prometheus monitoring objects.", "Error", err.Error())
		return ErrMetricsFailedRegisterPromCRDs
	}

	// Generate Service Object
	s, svcerr := GenerateService(8080, "metrics")
	if svcerr != nil {
		log.Info("Error generating metrics service object.", "Error", svcerr.Error())
		return ErrMetricsFailedGenerateService
	}
	log.Info("Generated metrics service object")

	// Generate ServiceMonitor Object
	sm := GenerateServiceMonitor(s)
	log.Info("Generated metrics servicemonitor object")

	// Create or Update Service
	err := mgr.GetClient().Create(context.TODO(), s)
	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			// Update the service if it already exists
			if updateErr := mgr.GetClient().Update(context.TODO(), s); updateErr != nil {
				log.Info("Error creating metrics service", "Error", updateErr.Error())
				return ErrMetricsFailedCreateService
			}
			log.Info("Error creating metrics service", "Error", err.Error())
			return ErrMetricsFailedCreateService
		}
	}
	log.Info("Created Service")

	// Create or Update the ServiceMonitor
	err = mgr.GetClient().Create(context.TODO(), sm)
	if err != nil {
		if k8serr.IsAlreadyExists(err) {
			// update the servicemonitor
			if smUpdateErr := mgr.GetClient().Update(context.TODO(), sm); smUpdateErr != nil {
				log.Info("Error creating metrics servicemonitor", "Error", smUpdateErr.Error())
				return ErrMetricsFailedCreateServiceMonitor
			}
		}
		log.Info("Error creating metrics servicemonitor", "Error", err.Error())
		return ErrMetricsFailedCreateServiceMonitor

	}
	log.Info("Created ServiceMonitor")

	return nil
}
