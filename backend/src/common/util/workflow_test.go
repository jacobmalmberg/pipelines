// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"testing"

	workflowapi "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	swfapi "github.com/kubeflow/pipelines/backend/src/crd/pkg/apis/scheduledworkflow/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestWorkflow_NewWorkflowFromBytes(t *testing.T) {
	// Error case
	workflow, err := NewWorkflowFromBytes([]byte("this is invalid format"))
	assert.Empty(t, workflow)
	assert.Error(t, err)
	assert.EqualError(t, err,
		"InvalidInputError: Failed to unmarshal the inputs: "+
			"error unmarshaling JSON: while decoding JSON: json: cannot unmarshal "+
			"string into Go value of type v1alpha1.Workflow")

	// Normal case
	bytes, err := yaml.Marshal(workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "WORKFLOW_NAME",
			Labels: map[string]string{"key": "value"},
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
		Status: workflowapi.WorkflowStatus{
			Message: "I AM A MESSAGE",
		},
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, bytes)

	workflow, err = NewWorkflowFromBytes(bytes)
	assert.Empty(t, err)
	assert.NotEmpty(t, workflow)
}

func TestWorkflow_NewWorkflowFromInterface(t *testing.T) {
	// Error case
	workflow, err := NewWorkflowFromInterface("this is invalid format")
	assert.Empty(t, workflow)
	assert.Error(t, err)
	assert.EqualError(t, err,
		NewInvalidInputError("not Workflow struct").Error())

	// Normal case
	workflow, err = NewWorkflowFromInterface(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "WORKFLOW_NAME",
			Labels: map[string]string{"key": "value"},
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
		Status: workflowapi.WorkflowStatus{
			Message: "I AM A MESSAGE",
		},
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, workflow)
}

func TestWorkflow_ScheduledWorkflowUUIDAsStringOrEmpty(t *testing.T) {
	// Base case
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "kubeflow.org/v1beta1",
				Kind:       "ScheduledWorkflow",
				Name:       "SCHEDULE_NAME",
				UID:        types.UID("MY_UID"),
			}},
		},
	})
	assert.Equal(t, "MY_UID", workflow.ScheduledWorkflowUUIDAsStringOrEmpty())
	assert.Equal(t, true, workflow.HasScheduledWorkflowAsParent())

	// No kind
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "kubeflow.org/v1beta1",
				UID:        types.UID("MY_UID"),
			}},
		},
	})
	assert.Equal(t, "", workflow.ScheduledWorkflowUUIDAsStringOrEmpty())
	assert.Equal(t, false, workflow.HasScheduledWorkflowAsParent())

	// Wrong kind
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "kubeflow.org/v1beta1",
				Kind:       "WRONG_KIND",
				Name:       "SCHEDULE_NAME",
				UID:        types.UID("MY_UID"),
			}},
		},
	})
	assert.Equal(t, "", workflow.ScheduledWorkflowUUIDAsStringOrEmpty())
	assert.Equal(t, false, workflow.HasScheduledWorkflowAsParent())

	// No API version
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				Kind: "ScheduledWorkflow",
				Name: "SCHEDULE_NAME",
				UID:  types.UID("MY_UID"),
			}},
		},
	})
	assert.Equal(t, "", workflow.ScheduledWorkflowUUIDAsStringOrEmpty())
	assert.Equal(t, false, workflow.HasScheduledWorkflowAsParent())

	// No UID
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "kubeflow.org/v1beta1",
				Kind:       "ScheduledWorkflow",
				Name:       "SCHEDULE_NAME",
			}},
		},
	})
	assert.Equal(t, "", workflow.ScheduledWorkflowUUIDAsStringOrEmpty())
	assert.Equal(t, false, workflow.HasScheduledWorkflowAsParent())

}

func TestWorkflow_ScheduledAtInSecOr0(t *testing.T) {
	// Base case
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			Labels: map[string]string{
				"scheduledworkflows.kubeflow.org/isOwnedByScheduledWorkflow": "true",
				"scheduledworkflows.kubeflow.org/scheduledWorkflowName":      "SCHEDULED_WORKFLOW_NAME",
				"scheduledworkflows.kubeflow.org/workflowEpoch":              "100",
				"scheduledworkflows.kubeflow.org/workflowIndex":              "50"},
		},
	})
	assert.Equal(t, int64(100), workflow.ScheduledAtInSecOr0())

	// No scheduled epoch
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			Labels: map[string]string{
				"scheduledworkflows.kubeflow.org/isOwnedByScheduledWorkflow": "true",
				"scheduledworkflows.kubeflow.org/scheduledWorkflowName":      "SCHEDULED_WORKFLOW_NAME",
				"scheduledworkflows.kubeflow.org/workflowIndex":              "50"},
		},
	})
	assert.Equal(t, int64(0), workflow.ScheduledAtInSecOr0())

	// No map
	workflow = NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
	})
	assert.Equal(t, int64(0), workflow.ScheduledAtInSecOr0())
}

func TestCondition(t *testing.T) {
	// Base case
	workflow := NewWorkflow(&workflowapi.Workflow{
		Status: workflowapi.WorkflowStatus{
			Phase: workflowapi.WorkflowRunning,
		},
	})
	assert.Equal(t, "Running", workflow.Condition())

	// No status
	workflow = NewWorkflow(&workflowapi.Workflow{
		Status: workflowapi.WorkflowStatus{},
	})
	assert.Equal(t, "", workflow.Condition())
}

func TestToStringForStore(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
	})
	assert.Equal(t,
		"{\"metadata\":{\"name\":\"WORKFLOW_NAME\",\"creationTimestamp\":null},\"spec\":{\"arguments\":{}},\"status\":{\"startedAt\":null,\"finishedAt\":null}}",
		workflow.ToStringForStore())
}

func TestWorkflow_OverrideName(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
	})

	workflow.OverrideName("NEW_WORKFLOW_NAME")

	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "NEW_WORKFLOW_NAME",
		},
	}

	assert.Equal(t, expected, workflow.Get())
}

func TestWorkflow_OverrideParameters(t *testing.T) {
	var tests = []struct {
		name      string
		workflow  *workflowapi.Workflow
		overrides map[string]string
		expected  *workflowapi.Workflow
	}{
		{
			name: "override parameters",
			workflow: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "WORKFLOW_NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1", Value: workflowapi.AnyStringPtr("VALUE1")},
							{Name: "PARAM2", Value: workflowapi.AnyStringPtr("VALUE2")},
							{Name: "PARAM3", Value: workflowapi.AnyStringPtr("VALUE3")},
							{Name: "PARAM4", Value: workflowapi.AnyStringPtr("")},
							{Name: "PARAM5", Value: workflowapi.AnyStringPtr("VALUE5")},
						},
					},
				},
			},
			overrides: map[string]string{
				"PARAM1": "NEW_VALUE1",
				"PARAM3": "NEW_VALUE3",
				"PARAM4": "NEW_VALUE4",
				"PARAM5": "",
				"PARAM9": "NEW_VALUE9",
			},
			expected: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "WORKFLOW_NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1", Value: workflowapi.AnyStringPtr("NEW_VALUE1")},
							{Name: "PARAM2", Value: workflowapi.AnyStringPtr("VALUE2")},
							{Name: "PARAM3", Value: workflowapi.AnyStringPtr("NEW_VALUE3")},
							{Name: "PARAM4", Value: workflowapi.AnyStringPtr("NEW_VALUE4")},
							{Name: "PARAM5", Value: workflowapi.AnyStringPtr("")},
						},
					},
				},
			},
		},
		{
			name: "handles missing parameter values",
			workflow: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1"}, // note, there's no value here
						},
					},
				},
			},
			overrides: nil,
			expected: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1"},
						},
					},
				},
			},
		},
		{
			name: "overrides a missing parameter value",
			workflow: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1"}, // note, there's no value here
						},
					},
				},
			},
			overrides: map[string]string{
				"PARAM1": "VALUE1",
			},
			expected: &workflowapi.Workflow{
				ObjectMeta: metav1.ObjectMeta{
					Name: "NAME",
				},
				Spec: workflowapi.WorkflowSpec{
					Arguments: workflowapi.Arguments{
						Parameters: []workflowapi.Parameter{
							{Name: "PARAM1", Value: workflowapi.AnyStringPtr("VALUE1")},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := NewWorkflow(tt.workflow)
			workflow.OverrideParameters(tt.overrides)
			assert.Equal(t, tt.expected, workflow.Get())
		})
	}
}

func TestWorkflow_SetOwnerReferences(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
	})

	workflow.SetOwnerReferences(&swfapi.ScheduledWorkflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "SCHEDULE_NAME",
		},
	})

	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion:         "kubeflow.org/v1beta1",
				Kind:               "ScheduledWorkflow",
				Name:               "SCHEDULE_NAME",
				Controller:         BoolPointer(true),
				BlockOwnerDeletion: BoolPointer(true),
			}},
		},
	}

	assert.Equal(t, expected, workflow.Get())
}

func TestWorkflow_SetLabelsToAllTemplates(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
		Spec: workflowapi.WorkflowSpec{
			Templates: []workflowapi.Template{
				{Metadata: workflowapi.Metadata{}},
			},
		},
	})
	workflow.SetLabelsToAllTemplates("key", "value")
	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
		Spec: workflowapi.WorkflowSpec{
			Templates: []workflowapi.Template{{
				Metadata: workflowapi.Metadata{
					Labels: map[string]string{
						"key": "value",
					},
				},
			}},
		},
	}

	assert.Equal(t, expected, workflow.Get())
}

func TestSetLabels(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
	})

	workflow.SetLabels("key", "value")

	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "WORKFLOW_NAME",
			Labels: map[string]string{"key": "value"},
		},
	}

	assert.Equal(t, expected, workflow.Get())
}

func TestGetWorkflowSpec(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "WORKFLOW_NAME",
			Labels: map[string]string{"key": "value"},
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
		Status: workflowapi.WorkflowStatus{
			Message: "I AM A MESSAGE",
		},
	})

	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "WORKFLOW_NAME",
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
	}

	assert.Equal(t, expected, workflow.GetWorkflowSpec().Get())
}

func TestGetWorkflowSpecTruncatesNameIfLongerThan200Runes(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "THIS_NAME_IS_GREATER_THAN_200_CHARACTERS_AND_WILL_BE_TRUNCATED_AFTER_THE_X_OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOXZZZZZZZZ",
			Labels: map[string]string{"key": "value"},
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
		Status: workflowapi.WorkflowStatus{
			Message: "I AM A MESSAGE",
		},
	})

	expected := &workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "THIS_NAME_IS_GREATER_THAN_200_CHARACTERS_AND_WILL_BE_TRUNCATED_AFTER_THE_X_OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOX",
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM", Value: workflowapi.AnyStringPtr("VALUE")},
				},
			},
		},
	}

	assert.Equal(t, expected, workflow.GetWorkflowSpec().Get())
}

func TestVerifyParameters(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM1", Value: workflowapi.AnyStringPtr("NEW_VALUE1")},
					{Name: "PARAM2", Value: workflowapi.AnyStringPtr("VALUE2")},
					{Name: "PARAM3", Value: workflowapi.AnyStringPtr("NEW_VALUE3")},
					{Name: "PARAM5", Value: workflowapi.AnyStringPtr("")},
				},
			},
		},
	})
	assert.Nil(t, workflow.VerifyParameters(map[string]string{"PARAM1": "V1", "PARAM2": "V2"}))
}

func TestVerifyParameters_Failed(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: "WORKFLOW_NAME",
		},
		Spec: workflowapi.WorkflowSpec{
			Arguments: workflowapi.Arguments{
				Parameters: []workflowapi.Parameter{
					{Name: "PARAM1", Value: workflowapi.AnyStringPtr("NEW_VALUE1")},
					{Name: "PARAM2", Value: workflowapi.AnyStringPtr("VALUE2")},
					{Name: "PARAM3", Value: workflowapi.AnyStringPtr("NEW_VALUE3")},
					{Name: "PARAM5", Value: workflowapi.AnyStringPtr("")},
				},
			},
		},
	})
	assert.NotNil(t, workflow.VerifyParameters(map[string]string{"PARAM1": "V1", "NON_EXIST": "V2"}))
}

func TestFindS3ArtifactKey_Succeed(t *testing.T) {
	expectedPath := "expected/path"
	workflow := NewWorkflow(&workflowapi.Workflow{
		Status: workflowapi.WorkflowStatus{
			Nodes: map[string]workflowapi.NodeStatus{
				"node-1": {
					Outputs: &workflowapi.Outputs{
						Artifacts: []workflowapi.Artifact{{
							Name: "artifact-1",
							ArtifactLocation: workflowapi.ArtifactLocation{
								S3: &workflowapi.S3Artifact{
									Key: expectedPath,
								},
							},
						}},
					},
				},
			},
		},
	})

	actualPath := workflow.FindObjectStoreArtifactKeyOrEmpty("node-1", "artifact-1")

	assert.Equal(t, expectedPath, actualPath)
}

func TestFindS3ArtifactKey_ArtifactNotFound(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		Status: workflowapi.WorkflowStatus{
			Nodes: map[string]workflowapi.NodeStatus{
				"node-1": {
					Outputs: &workflowapi.Outputs{
						Artifacts: []workflowapi.Artifact{{
							Name: "artifact-2",
							ArtifactLocation: workflowapi.ArtifactLocation{
								S3: &workflowapi.S3Artifact{
									Key: "foo/bar",
								},
							},
						}},
					},
				},
			},
		},
	})

	actualPath := workflow.FindObjectStoreArtifactKeyOrEmpty("node-1", "artifact-1")

	assert.Empty(t, actualPath)
}

func TestFindS3ArtifactKey_NodeNotFound(t *testing.T) {
	workflow := NewWorkflow(&workflowapi.Workflow{
		Status: workflowapi.WorkflowStatus{
			Nodes: map[string]workflowapi.NodeStatus{},
		},
	})

	actualPath := workflow.FindObjectStoreArtifactKeyOrEmpty("node-1", "artifact-1")

	assert.Empty(t, actualPath)
}

func TestReplaceUID(t *testing.T) {
	workflowString := `apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: k8s-owner-reference-
spec:
  entrypoint: k8s-owner-reference
  templates:
  - name: k8s-owner-reference
    resource:
      action: create
      manifest: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          generateName: owned-eg-
          ownerReferences:
          - apiVersion: argoproj.io/v1alpha1
            blockOwnerDeletion: true
            kind: Workflow
            name: "{{workflow.name}}"
            uid: "{{workflow.uid}}"
        data:
          some: value`
	var workflow Workflow
	err := yaml.Unmarshal([]byte(workflowString), &workflow)
	assert.Nil(t, err)
	workflow.ReplaceUID("12345")
	expectedWorkflowString := `apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: k8s-owner-reference-
spec:
  entrypoint: k8s-owner-reference
  templates:
  - name: k8s-owner-reference
    resource:
      action: create
      manifest: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          generateName: owned-eg-
          ownerReferences:
          - apiVersion: argoproj.io/v1alpha1
            blockOwnerDeletion: true
            kind: Workflow
            name: "{{workflow.name}}"
            uid: "12345"
        data:
          some: value`

	var expectedWorkflow Workflow
	err = yaml.Unmarshal([]byte(expectedWorkflowString), &expectedWorkflow)
	assert.Nil(t, err)
	assert.Equal(t, expectedWorkflow, workflow)
}
