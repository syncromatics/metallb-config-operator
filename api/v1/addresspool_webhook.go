/*

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

package v1

import (
	"net"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var addresspoollog = logf.Log.WithName("addresspool-resource")

func (r *AddressPool) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-app-metallb-com-v1-addresspool,mutating=true,failurePolicy=fail,groups=app.metallb.com,resources=addresspools,verbs=create;update,versions=v1,name=maddresspool.kb.io

var _ webhook.Defaulter = &AddressPool{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AddressPool) Default() {
	addresspoollog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-app-metallb-com-v1-addresspool,mutating=false,failurePolicy=fail,groups=app.metallb.com,resources=addresspools,versions=v1,name=vaddresspool.kb.io

var _ webhook.Validator = &AddressPool{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AddressPool) ValidateCreate() error {
	addresspoollog.Info("validate create", "name", r.Name)

	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AddressPool) ValidateUpdate(old runtime.Object) error {
	addresspoollog.Info("validate update", "name", r.Name)

	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AddressPool) ValidateDelete() error {
	addresspoollog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *AddressPool) validate() error {
	switch r.Spec.Protocol {
	case "layer2":
	case "bgp":
	default:
		return errors.Errorf("unkown protocol type '%s', supported types are ['layer2','bgp']", r.Spec.Protocol)
	}

	if len(r.Spec.Addresses) == 0 {
		return errors.Errorf("you must define at least one address")
	}

	for _, a := range r.Spec.Addresses {
		_, _, err := net.ParseCIDR(a)
		if err != nil {
			return errors.Wrapf(err, "'%s' is an invalid CIDR", a)
		}
	}
	return nil
}
