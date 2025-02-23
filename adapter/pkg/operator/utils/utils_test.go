/*
 *  Copyright (c) 2023, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	dpv1alpha1 "github.com/wso2/apk/adapter/pkg/operator/apis/dp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTieBreaker(t *testing.T) {

	type testItem struct {
		objectList     []*dpv1alpha1.BackendPolicy
		expectedObject *dpv1alpha1.BackendPolicy
		message        string
	}

	newTime := time.Now()
	newTimePlusOneMinute := newTime.Add(time.Minute * time.Duration(1))

	policy1 := dpv1alpha1.BackendPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "policy-1",
			CreationTimestamp: metav1.NewTime(newTime),
		},
		Spec: dpv1alpha1.BackendPolicySpec{
			Default: &dpv1alpha1.BackendConfigs{
				Protocol: dpv1alpha1.HTTPProtocol,
			},
		},
	}

	policy2 := dpv1alpha1.BackendPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "policy-2",
			CreationTimestamp: metav1.NewTime(newTimePlusOneMinute),
		},
		Spec: dpv1alpha1.BackendPolicySpec{
			Default: &dpv1alpha1.BackendConfigs{
				Protocol: dpv1alpha1.HTTPProtocol,
			},
		},
	}

	policy3 := dpv1alpha1.BackendPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "default",
			Name:              "policy-0",
			CreationTimestamp: metav1.NewTime(newTime),
		},
		Spec: dpv1alpha1.BackendPolicySpec{
			Default: &dpv1alpha1.BackendConfigs{
				Protocol: dpv1alpha1.HTTPProtocol,
			},
		},
	}

	tests := []testItem{
		{
			objectList:     []*dpv1alpha1.BackendPolicy{&policy1, &policy2},
			expectedObject: &policy1,
			message:        "Tie breaking using creation timestamps are different is not working",
		},
		{
			objectList:     []*dpv1alpha1.BackendPolicy{&policy1, &policy3},
			expectedObject: &policy3,
			message:        "Tie breaking using creation timestamps are equal is not working",
		},
	}

	for _, test := range tests {
		actualOutput := TieBreaker(test.objectList)
		assert.Equal(t, test.expectedObject, *actualOutput, test.message)
	}
}
