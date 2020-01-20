package harbor

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
)

func (r *Reconciler) GetCondition(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, conditionType containerregistryv1alpha1.HarborConditionType) containerregistryv1alpha1.HarborCondition {
	for _, condition := range harbor.Status.Conditions {
		if condition.Type == conditionType {
			return condition
		}
	}

	return containerregistryv1alpha1.HarborCondition{
		Type:   conditionType,
		Status: corev1.ConditionUnknown,
	}
}

func (r *Reconciler) GetConditionStatus(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, conditionType containerregistryv1alpha1.HarborConditionType) corev1.ConditionStatus {
	return r.GetCondition(ctx, harbor, conditionType).Status
}

func (r *Reconciler) UpdateCondition(ctx context.Context, harbor *containerregistryv1alpha1.Harbor, conditionType containerregistryv1alpha1.HarborConditionType, status corev1.ConditionStatus, reasons ...string) error {
	var reason, message string

	switch len(reasons) {
	case 0: // nolint:mnd
	case 1: // nolint:mnd
		reason = reasons[0]
	case 2: // nolint:mnd
		reason = reasons[0]
		message = reasons[1]
	default:
		return errors.Errorf("expecting reason and message, got %d parameters", len(reasons))
	}

	now := metav1.Now()

	for i, condition := range harbor.Status.Conditions {
		if condition.Type == conditionType {
			now.DeepCopyInto(&condition.LastUpdateTime)

			if condition.LastTransitionTime.IsZero() || condition.Status != status {
				now.DeepCopyInto(&condition.LastTransitionTime)
			}

			condition.Status = status
			condition.Reason = reason
			condition.Message = message

			harbor.Status.Conditions[i] = condition

			return nil
		}
	}

	condition := containerregistryv1alpha1.HarborCondition{
		Type:    conditionType,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
	now.DeepCopyInto(&condition.LastUpdateTime)
	now.DeepCopyInto(&condition.LastTransitionTime)

	harbor.Status.Conditions = append(harbor.Status.Conditions, condition)

	return nil
}

// UpdateStatus applies current in-memory statuses to the remote resource
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#status-subresource
func (r *Reconciler) UpdateStatus(ctx context.Context, result *ctrl.Result, harbor *containerregistryv1alpha1.Harbor) error {
	err := r.Status().Update(ctx, harbor)
	if err != nil {
		result.Requeue = true

		return errors.Wrap(err, "cannot update status field")
	}

	return nil
}
