// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package controllers

import (
	"context"
	"errors"
	"testing"

	mock_condition "github.com/aws/amazon-vpc-resource-controller-k8s/mocks/amazon-vcp-resource-controller-k8s/pkg/condition"
	mock_node "github.com/aws/amazon-vpc-resource-controller-k8s/mocks/amazon-vcp-resource-controller-k8s/pkg/node"
	mock_manager "github.com/aws/amazon-vpc-resource-controller-k8s/mocks/amazon-vcp-resource-controller-k8s/pkg/node/manager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	mockConfigMapName = "amazon-vpc-cni"
	mockConfigMapNS   = "configmap-ns"

	mockConfigMap = &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockConfigMapName,
			Namespace: mockConfigMapNS,
		},
	}
	mockConfigMapReq = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: mockConfigMapNS,
			Name:      mockConfigMapName,
		},
	}
	mockErr = errors.New("Mock error")
)

type ConfigMapMock struct {
	MockNodeManager     *mock_manager.MockManager
	MockConditions      *mock_condition.MockConditions
	MockNode            *mock_node.MockNode
	ConfigMapReconciler *ConfigMapReconciler
}

func NewConfigMapMock(ctrl *gomock.Controller, mockObjects ...runtime.Object) ConfigMapMock {
	mockNodeManager := mock_manager.NewMockManager(ctrl)
	mockConditions := mock_condition.NewMockConditions(ctrl)
	mockNode := mock_node.NewMockNode(ctrl)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	client := fakeClient.NewFakeClientWithScheme(scheme, mockObjects...)

	return ConfigMapMock{
		MockNodeManager: mockNodeManager,
		ConfigMapReconciler: &ConfigMapReconciler{
			Client:      client,
			Log:         zap.New(),
			NodeManager: mockNodeManager,
		},
		MockNode:       mockNode,
		MockConditions: mockConditions,
	}
}

func Test_Reconcile_UpdateNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewConfigMapMock(ctrl, mockConfigMap)

	// mock.MockConditions.EXPECT().IsWindowsIPAMEnabled().Return(true)
	mock.MockNodeManager.EXPECT().UpdateNodesOnConfigMapChanges().Return(nil)

	res, err := mock.ConfigMapReconciler.Reconcile(context.TODO(), mockConfigMapReq)
	assert.NoError(t, err)
	assert.Equal(t, res, reconcile.Result{})
}

func Test_Reconcile_UpdateNode_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewConfigMapMock(ctrl, mockConfigMap)

	// mock.MockConditions.EXPECT().IsWindowsIPAMEnabled().Return(false)
	mock.MockNodeManager.EXPECT().UpdateNodesOnConfigMapChanges().Return(mockErr)

	res, err := mock.ConfigMapReconciler.Reconcile(context.TODO(), mockConfigMapReq)
	assert.Error(t, err)
	assert.Equal(t, res, reconcile.Result{})
}
