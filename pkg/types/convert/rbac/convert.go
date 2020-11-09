package rbac

import (
	"github.com/rancher/dolly/pkg/dollyfile"
	"github.com/rancher/dolly/pkg/types"
	"github.com/rancher/dolly/pkg/types/convert/labels"
	"github.com/rancher/dolly/pkg/types/utils"
	"github.com/rancher/wrangler/pkg/name"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Plugin struct{}

func (p Plugin) Convert(rf *dollyfile.DollyFile) (result []runtime.Object) {
	for _, service := range rf.Services {
		labels := labels.SelectorLabels(service)
		subject := subject(service)
		if subject == nil {
			return nil
		}

		var result []runtime.Object

		// serviceAccount
		result = append(result, serviceAccount(labels, *subject))

		// role and rolebindings
		result = append(result, roles(*subject, service, labels)...)
		result = append(result, rules(*subject, service, labels)...)

		// clusterrole and clusterrolebindings
		result = append(result, clusterRoles(service, *subject, labels)...)
		result = append(result, clusterRules(service, *subject, labels)...)
	}

	return result
}

func clusterRules(service types.Service, subject rbacv1.Subject, labels map[string]string) []runtime.Object {
	var result []runtime.Object
	role := NewClusterRole(service.Name, labels)
	for _, perm := range service.Spec.GlobalPermissions {
		if perm.Role != "" {
			continue
		}
		policyRule, ok := PermToPolicyRule(perm)
		if ok {
			role.Rules = append(role.Rules, policyRule)
		}
	}

	if len(role.Rules) > 0 {
		result = append(result, role)

		roleBinding := NewClusterBinding(name.SafeConcatName(service.Name, role.Name), labels)
		roleBinding.Subjects = []rbacv1.Subject{
			subject,
		}
		roleBinding.RoleRef = rbacv1.RoleRef{
			Name:     role.Name,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		}
		result = append(result, roleBinding)

	}
	return result
}

func clusterRoles(service types.Service, subject rbacv1.Subject, labels map[string]string) []runtime.Object {
	var result []runtime.Object
	for _, role := range service.Spec.GlobalPermissions {
		if role.Role == "" {
			continue
		}
		roleBinding := NewClusterBinding(name.SafeConcatName("cluster", service.Name, role.Role), labels)
		roleBinding.Subjects = []rbacv1.Subject{
			subject,
		}
		roleBinding.RoleRef = rbacv1.RoleRef{
			Name:     role.Role,
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		}
		result = append(result, roleBinding)
	}
	return result
}

func subject(service types.Service) *rbacv1.Subject {
	name := utils.ServiceAccountName(service)
	if name == "" {
		return nil
	}

	return &rbacv1.Subject{
		Name:      name,
		Namespace: service.Namespace,
		Kind:      "ServiceAccount",
	}
}

func roles(subject rbacv1.Subject, service types.Service, labels map[string]string) []runtime.Object {
	var result []runtime.Object
	for _, role := range service.Spec.Permissions {
		if role.Role == "" {
			continue
		}
		roleBinding := NewBinding(service.Namespace, name.SafeConcatName(service.Name, role.Role), labels)
		roleBinding.Subjects = []rbacv1.Subject{
			subject,
		}
		roleBinding.RoleRef = rbacv1.RoleRef{
			Name:     role.Role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		result = append(result, roleBinding)
	}
	return result
}

func rules(subject rbacv1.Subject, service types.Service, labels map[string]string) []runtime.Object {
	var result []runtime.Object
	role := NewRole(service.Namespace, name.SafeConcatName(service.Name), labels)
	for _, perm := range service.Spec.Permissions {
		if perm.Role != "" {
			continue
		}
		policyRule, ok := PermToPolicyRule(perm)
		if ok {
			role.Rules = append(role.Rules, policyRule)
		}
	}

	if len(role.Rules) > 0 {
		result = append(result, role)

		roleBinding := NewBinding(service.Namespace, name.SafeConcatName(service.Name, role.Name), labels)
		roleBinding.Subjects = []rbacv1.Subject{
			subject,
		}
		roleBinding.RoleRef = rbacv1.RoleRef{
			Name:     role.Name,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		}
		result = append(result, roleBinding)
	}
	return result
}

func serviceAccount(labels map[string]string, subject rbacv1.Subject) *v1.ServiceAccount {
	sa := newServiceAccount(subject.Namespace, subject.Name, labels)
	return sa
}

func PermToPolicyRule(perm types.Permission) (rbacv1.PolicyRule, bool) {
	policyRule := rbacv1.PolicyRule{}
	valid := false

	if perm.Role != "" {
		return policyRule, valid
	}

	policyRule.Verbs = perm.Verbs
	if perm.URL == "" {
		if perm.ResourceName != "" {
			valid = true
			policyRule.ResourceNames = []string{perm.ResourceName}
		}

		policyRule.APIGroups = []string{perm.APIGroup}

		if perm.Resource != "" {
			valid = true
			policyRule.Resources = []string{perm.Resource}
		}
	} else {
		valid = true
		policyRule.NonResourceURLs = []string{perm.URL}
	}

	return policyRule, valid
}

func newServiceAccount(namespace, name string, labels map[string]string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewRole(namespace, name string, labels map[string]string) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewClusterRole(name string, labels map[string]string) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewClusterBinding(name string, labels map[string]string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}

func NewBinding(namespace, name string, labels map[string]string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: map[string]string{},
		},
	}
}
