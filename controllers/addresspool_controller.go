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

package controllers

import (
	"context"
	"fmt"
	"sort"

	appv1 "github.com/syncromatics/metallb-config-operator/api/v1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AddressPoolReconciler reconciles a AddressPool object
type AddressPoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	ConfigName      string
	ConfigNamespace string
}

// +kubebuilder:rbac:groups=app.metallb.com,resources=addresspools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.metallb.com,resources=addresspools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;update;patch;create;list;watch

func (r *AddressPoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("addresspool", req.NamespacedName)

	list := &appv1.AddressPoolList{}

	err := r.Client.List(context.Background(), list)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to get addresspools")
	}

	newPool := r.createPool(list)

	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.ConfigName,
			Namespace: r.ConfigNamespace,
		},
	}
	err = r.Client.Get(context.Background(), client.ObjectKey{
		Name:      r.ConfigName,
		Namespace: r.ConfigNamespace,
	}, config)
	if kerrors.IsNotFound(err) {
		return ctrl.Result{}, r.updateConfigMap(config, newPool, true)
	}
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to get configmap")
	}

	configYaml, ok := config.Data["config"]
	if !ok {
		return ctrl.Result{}, r.updateConfigMap(config, newPool, false)
	}

	configMap := &configFile{}
	err = yaml.Unmarshal([]byte(configYaml), configMap)
	if err != nil {
		return ctrl.Result{}, r.updateConfigMap(config, newPool, false)
	}

	update := r.isDifferent(configMap, newPool)
	if update {
		return ctrl.Result{}, r.updateConfigMap(config, newPool, false)
	}

	fmt.Println(config.Data["config"])

	return ctrl.Result{}, nil
}

func (r *AddressPoolReconciler) createPool(list *appv1.AddressPoolList) *configFile {
	config := &configFile{}

	for _, a := range list.Items {
		ip := addressPool{
			Name:      a.Name,
			Protocol:  a.Spec.Protocol,
			Addresses: a.Spec.Addresses,
		}

		config.Pools = append(config.Pools, ip)
	}

	return config
}

func (r *AddressPoolReconciler) isDifferent(pool *configFile, newPool *configFile) bool {
	if len(pool.Pools) != len(newPool.Pools) {
		return true
	}

	sort.SliceStable(pool.Pools, func(i, j int) bool {
		return pool.Pools[i].Name < pool.Pools[j].Name
	})

	sort.SliceStable(newPool.Pools, func(i, j int) bool {
		return newPool.Pools[i].Name < newPool.Pools[j].Name
	})

	for i := 0; i < len(pool.Pools); i++ {
		if pool.Pools[i].Name != newPool.Pools[i].Name {
			return true
		}
		if len(pool.Pools[i].Addresses) != len(newPool.Pools[i].Addresses) {
			return true
		}
		if pool.Pools[i].Protocol != newPool.Pools[i].Protocol {
			return true
		}

		for j := 0; j < len(pool.Pools[i].Addresses); j++ {
			if pool.Pools[i].Addresses[j] != newPool.Pools[i].Addresses[j] {
				return true
			}
		}
	}
	return false
}

func (r *AddressPoolReconciler) updateConfigMap(config *corev1.ConfigMap, pool *configFile, isCreate bool) error {
	b, err := yaml.Marshal(pool)
	if err != nil {
		return errors.Wrap(err, "failed to marshal yaml")
	}

	if config.Data == nil {
		config.Data = map[string]string{}
	}

	config.Data["config"] = string(b)

	if isCreate {
		err = r.Client.Create(context.Background(), config)
		if err != nil {
			return errors.Wrap(err, "failed to create configmap")
		}
		return nil
	}

	err = r.Client.Update(context.Background(), config)
	if err != nil {
		return errors.Wrap(err, "failed to update configmap")
	}

	return nil
}

func (r *AddressPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1.AddressPool{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

type configFile struct {
	Pools []addressPool `yaml:"address-pools"`
}

type addressPool struct {
	Protocol  string
	Name      string
	Addresses []string
}
