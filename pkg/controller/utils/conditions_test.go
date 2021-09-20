package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	//awsv1alpha1 "github.com/openshift/aws-account-operator/pkg/apis/aws/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

func TestShouldUpdateCondition(t *testing.T) {

	testData := []struct {
		name                 string
		oldStatus            corev1.ConditionStatus
		newStatus            corev1.ConditionStatus
		updateConditionCheck UpdateConditionCheck
		expectedReturn       bool
	}{
		{
			name:                 "test different old status and new status",
			oldStatus:            corev1.ConditionTrue,
			newStatus:            corev1.ConditionFalse,
			updateConditionCheck: UpdateConditionAlways,
			expectedReturn:       true,
		},
		{
			name:                 "test same old status and new status but always update conditions",
			oldStatus:            corev1.ConditionFalse,
			newStatus:            corev1.ConditionFalse,
			updateConditionCheck: UpdateConditionAlways,
			expectedReturn:       true,
		},
		{
			name:                 "test same old status and new status but never update conditions",
			oldStatus:            corev1.ConditionTrue,
			newStatus:            corev1.ConditionTrue,
			updateConditionCheck: UpdateConditionNever,
			expectedReturn:       false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {

			oldReason := "randOldReason"
			oldMessage := "random old msg"
			newReason := "randNewReason"
			newMessage := "random new message"

			returnVal := shouldUpdateCondition(test.oldStatus, oldReason, oldMessage,
				test.newStatus, newReason, newMessage, test.updateConditionCheck)
			assert.Equal(t, returnVal, test.expectedReturn)
		})
	}
}

// func TestSetAccountCondition(t *testing.T) {

// 	testData := []struct{
// 		name string
// 		oldStatus corev1.ConditionStatus
// 		newStatus corev1.ConditionStatus
// 		updateConditionCheck UpdateConditionCheck
// 		expectedReturn []awsv1alpha1.AccountCondition
// 	}{
// 		{
// 			name: "test different old status and new status",
// 			oldStatus: corev1.ConditionTrue,
// 			newStatus: corev1.ConditionFalse,
// 			updateConditionCheck: UpdateConditionAlways,
// 			expectedReturn: []awsv1alpha1.AccountCondition{},
// 		},
// 		{
// 			name: "test same old status and new status but always update conditions",
// 			oldStatus: corev1.ConditionFalse,
// 			newStatus: corev1.ConditionFalse,
// 			updateConditionCheck: UpdateConditionAlways,
// 			expectedReturn: []awsv1alpha1.AccountCondition{},
// 		},
// 		{
// 			name: "test same old status and new status but never update conditions",
// 			oldStatus: corev1.ConditionTrue,
// 			newStatus: corev1.ConditionTrue,
// 			updateConditionCheck: UpdateConditionNever,
// 			expectedReturn: []awsv1alpha1.AccountCondition{},
// 		},
// 	}

// 	for _, test := range testData {
// 		t.Run(test.name, func(t *testing.T){

// 			oldReason := "randOldReason"
// 			oldMessage := "random old msg"
// 			newReason := "randNewReason"
// 			newMessage := "random new message"
// 			ccs := true

// 			returnVal := SetAccountCondition()
// 		})
// 	}
// }
