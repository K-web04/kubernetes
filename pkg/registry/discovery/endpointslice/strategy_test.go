/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package endpointslice

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/apis/discovery"
	"k8s.io/kubernetes/pkg/features"
	utilpointer "k8s.io/utils/pointer"
)

func Test_dropDisabledFieldsOnCreate(t *testing.T) {
	testcases := []struct {
		name             string
		hintsGateEnabled bool
		eps              *discovery.EndpointSlice
		expectedEPS      *discovery.EndpointSlice
	}{
		{
			name: "node name gate enabled, field should be allowed",
			eps: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.TopologyAwareHints, testcase.hintsGateEnabled)()

			dropDisabledFieldsOnCreate(testcase.eps)
			if !apiequality.Semantic.DeepEqual(testcase.eps, testcase.expectedEPS) {
				t.Logf("actual endpointslice: %v", testcase.eps)
				t.Logf("expected endpointslice: %v", testcase.expectedEPS)
				t.Errorf("unexpected EndpointSlice on create API strategy")
			}
		})
	}
}

func Test_dropDisabledFieldsOnUpdate(t *testing.T) {
	testcases := []struct {
		name             string
		hintsGateEnabled bool
		oldEPS           *discovery.EndpointSlice
		newEPS           *discovery.EndpointSlice
		expectedEPS      *discovery.EndpointSlice
	}{
		{
			name: "node name gate enabled, set on new EPS",
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: nil,
					},
					{
						NodeName: nil,
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
		},
		{
			name: "node name gate disabled, set on old and updated EPS",
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1-old"),
					},
					{
						NodeName: utilpointer.String("node-2-old"),
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName: utilpointer.String("node-1"),
					},
					{
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
		},
		{
			name:             "hints gate enabled, set on new EPS",
			hintsGateEnabled: true,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: nil,
					},
					{
						Hints: nil,
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b"}},
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b"}},
						},
					},
				},
			},
		},
		{
			name:             "hints gate disabled, set on new EPS",
			hintsGateEnabled: false,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: nil,
					},
					{
						Hints: nil,
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b"}},
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: nil,
					},
					{
						Hints: nil,
					},
				},
			},
		},
		{
			name:             "hints gate disabled, set on new and old EPS",
			hintsGateEnabled: false,
			oldEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a-old"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b-old"}},
						},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b"}},
						},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-a"}},
						},
					},
					{
						Hints: &discovery.EndpointHints{
							ForZones: []discovery.ForZone{{Name: "zone-b"}},
						},
					},
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.TopologyAwareHints, testcase.hintsGateEnabled)()

			dropDisabledFieldsOnUpdate(testcase.oldEPS, testcase.newEPS)
			if !apiequality.Semantic.DeepEqual(testcase.newEPS, testcase.expectedEPS) {
				t.Logf("actual endpointslice: %v", testcase.newEPS)
				t.Logf("expected endpointslice: %v", testcase.expectedEPS)
				t.Errorf("unexpected EndpointSlice from update API strategy")
			}
		})
	}
}

func TestPrepareForUpdate(t *testing.T) {
	testCases := []struct {
		name        string
		oldEPS      *discovery.EndpointSlice
		newEPS      *discovery.EndpointSlice
		expectedEPS *discovery.EndpointSlice
	}{
		{
			name: "unchanged EPS should not increment generation",
			oldEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
			newEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
			expectedEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
		},
		{
			name: "changed endpoints should increment generation",
			oldEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
			newEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.5"},
				}},
			},
			expectedEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{Generation: 2},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.5"},
				}},
			},
		},
		{
			name: "changed labels should increment generation",
			oldEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Labels:     map[string]string{"example": "one"},
				},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
			newEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
					Labels:     map[string]string{"example": "two"},
				},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
			expectedEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
					Labels:     map[string]string{"example": "two"},
				},
				Endpoints: []discovery.Endpoint{{
					Addresses: []string{"1.2.3.4"},
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			Strategy.PrepareForUpdate(context.TODO(), tc.newEPS, tc.oldEPS)
			if !apiequality.Semantic.DeepEqual(tc.newEPS, tc.expectedEPS) {
				t.Errorf("Expected %+v\nGot: %+v", tc.expectedEPS, tc.newEPS)
			}
		})
	}
}

func Test_dropTopologyOnV1(t *testing.T) {
	testcases := []struct {
		name        string
		v1Request   bool
		newEPS      *discovery.EndpointSlice
		originalEPS *discovery.EndpointSlice
		expectedEPS *discovery.EndpointSlice
	}{
		{
			name:      "v1 request, without deprecated topology",
			v1Request: true,
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{Hostname: utilpointer.String("hostname-1")},
					{Hostname: utilpointer.String("hostname-1")},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{Hostname: utilpointer.String("hostname-1")},
					{Hostname: utilpointer.String("hostname-1")},
				},
			},
		},
		{
			name: "v1beta1 request, without deprecated topology",
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{Hostname: utilpointer.String("hostname-1")},
					{Hostname: utilpointer.String("hostname-1")},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{Hostname: utilpointer.String("hostname-1")},
					{Hostname: utilpointer.String("hostname-1")},
				},
			},
		},
		{
			name:      "v1 request, with deprecated topology",
			v1Request: true,
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"key": "value"}},
					{DeprecatedTopology: map[string]string{"key": "value"}},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{{}, {}},
			},
		},
		{
			name: "v1beta1 request, with deprecated topology",
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"key": "value"}},
					{DeprecatedTopology: map[string]string{"key": "value"}},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"key": "value"}},
					{DeprecatedTopology: map[string]string{"key": "value"}},
				},
			},
		},
		{
			name:      "v1 request, updated metadata",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"example": "one"},
				},
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"example": "one"},
				},
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name: "v1beta1 request, updated metadata",
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"example": "one"},
				},
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"example": "one"},
				},
				Endpoints: []discovery.Endpoint{
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
					{
						NodeName:           utilpointer.String("node-1"),
						DeprecatedTopology: map[string]string{"key": "value"},
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"key": "value"}},
					{DeprecatedTopology: map[string]string{"key": "value"}},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{Hostname: utilpointer.String("hostname-1")},
					{Hostname: utilpointer.String("hostname-1")},
				},
			},
		},
		{
			name: "v1beta1 request, updated endpoints",
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"key": "value"}},
					{DeprecatedTopology: map[string]string{"key": "value"}},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with topology node names + other topology fields",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1", "other": "value"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1", "foo": "bar"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1a"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1", "other": "value"},
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1", "foo": "bar"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						NodeName: utilpointer.String("node-1"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-1"),
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with topology node names",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1a"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						NodeName: utilpointer.String("node-1"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-1"),
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with topology node names swapped",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-2"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1a"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-2"},
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						NodeName: utilpointer.String("node-2"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-1"),
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with new topology node name",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-2"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						// Invalid node name because it did not exist in previous version of EndpointSlice
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-3"},
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-1"),
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with topology node names + 1 new node name",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1a"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						NodeName:           utilpointer.String("node-2"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						NodeName: utilpointer.String("node-1"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
		},
		{
			name:      "v1 request, updated endpoints with topology node names + new node names",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1a"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
						NodeName:           utilpointer.String("node-1"),
					},
					{
						Hostname:           utilpointer.String("hostname-1b"),
						NodeName:           utilpointer.String("node-2"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"},
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1a"),
						NodeName: utilpointer.String("node-1"),
					},
					{
						Hostname: utilpointer.String("hostname-1b"),
						NodeName: utilpointer.String("node-2"),
					},
				},
			},
		},
		{
			name:      "v1 request, invalid node name label",
			v1Request: true,
			originalEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "valid-node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-2"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "invalid node-2"},
					},
					{
						Hostname:           utilpointer.String("hostname-3"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "valid-node-3"},
					},
					{
						Hostname:           utilpointer.String("hostname-4"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "invalid node-4"},
					},
				},
			},
			newEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname:           utilpointer.String("hostname-1"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "valid-node-1"},
					},
					{
						Hostname:           utilpointer.String("hostname-2"),
						DeprecatedTopology: map[string]string{corev1.LabelHostname: "invalid node-2"},
					},
					{
						Hostname: utilpointer.String("hostname-3"),
						NodeName: utilpointer.String("node-3"),
					},
					{
						Hostname: utilpointer.String("hostname-4"),
						NodeName: utilpointer.String("node-4"),
					},
				},
			},
			expectedEPS: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{
						Hostname: utilpointer.String("hostname-1"),
						NodeName: utilpointer.String("valid-node-1"),
					},
					{
						Hostname: utilpointer.String("hostname-2"),
					},
					{
						Hostname: utilpointer.String("hostname-3"),
						NodeName: utilpointer.String("node-3"),
					},
					{
						Hostname: utilpointer.String("hostname-4"),
						NodeName: utilpointer.String("node-4"),
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := genericapirequest.WithRequestInfo(genericapirequest.NewContext(), &genericapirequest.RequestInfo{APIGroup: "discovery.k8s.io", APIVersion: "v1beta1", Resource: "endpointslices"})
			if tc.v1Request {
				ctx = genericapirequest.WithRequestInfo(genericapirequest.NewContext(), &genericapirequest.RequestInfo{APIGroup: "discovery.k8s.io", APIVersion: "v1", Resource: "endpointslices"})
			}

			dropTopologyOnV1(ctx, tc.originalEPS, tc.newEPS)
			if !apiequality.Semantic.DeepEqual(tc.newEPS, tc.expectedEPS) {
				t.Logf("actual endpointslice: %v", tc.newEPS)
				t.Logf("expected endpointslice: %v", tc.expectedEPS)
				t.Errorf("unexpected EndpointSlice on API topology strategy")
			}
		})
	}
}

func Test_getDeprecatedTopologyNodeNames(t *testing.T) {
	testcases := []struct {
		name              string
		endpointSlice     *discovery.EndpointSlice
		expectedNodeNames sets.String
	}{
		{
			name: "2 nodes",
			endpointSlice: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"}},
					{DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-2"}},
				},
			},
			expectedNodeNames: sets.NewString("node-1", "node-2"),
		},
		{
			name: "duplicate values",
			endpointSlice: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-1"}},
					{DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-3"}},
					{DeprecatedTopology: map[string]string{corev1.LabelHostname: "node-3"}},
				},
			},
			expectedNodeNames: sets.NewString("node-1", "node-3"),
		},
		{
			name: "unset",
			endpointSlice: &discovery.EndpointSlice{
				Endpoints: []discovery.Endpoint{
					{DeprecatedTopology: map[string]string{"other": "value"}},
					{DeprecatedTopology: map[string]string{"foo": "bar"}},
					{DeprecatedTopology: nil},
				},
			},
			expectedNodeNames: sets.NewString(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actualNames := getDeprecatedTopologyNodeNames(tc.endpointSlice)
			if !tc.expectedNodeNames.Equal(actualNames) {
				t.Errorf("Expected %+v node names, got %+v", tc.expectedNodeNames, actualNames)
			}
		})
	}
}
