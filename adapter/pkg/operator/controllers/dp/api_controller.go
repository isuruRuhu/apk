/*
 *  Copyright (c) 2022, WSO2 LLC. (http://www.wso2.org) All Rights Reserved.
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

package controllers

import (
	"context"
	"fmt"

	"github.com/wso2/apk/adapter/config"
	"github.com/wso2/apk/adapter/internal/loggers"
	"github.com/wso2/apk/adapter/pkg/logging"
	"github.com/wso2/apk/adapter/pkg/operator/constants"
	"github.com/wso2/apk/adapter/pkg/operator/status"
	"github.com/wso2/apk/adapter/pkg/operator/synchronizer"
	"github.com/wso2/apk/adapter/pkg/operator/utils"
	k8error "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
	k8client "sigs.k8s.io/controller-runtime/pkg/client"

	dpv1alpha1 "github.com/wso2/apk/adapter/pkg/operator/apis/dp/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	httpRouteAPIIndex = "httpRouteAPIIndex"
	// Index for API level authentications
	httpRouteAuthenticationIndex = "httpRouteAuthenticationIndex"
	// Index for resource level authentications
	httpRouteAuthenticationResourceIndex = "httpRouteAuthenticationResourceIndex"
	// Index for API level apipolicies
	httpRouteAPIPolicyIndex = "httpRouteAPIPolicyIndex"
	// Index for resource level apipolicies
	httpRouteAPIPolicyResourceIndex = "httpRouteAPIPolicyResourceIndex"
	serviceHTTPRouteIndex           = "serviceHTTPRouteIndex"
	serviceBackendPolicyIndex       = "serviceBackendPolicyIndex"
	apiScopeIndex                   = "apiScopeIndex"
)

// APIReconciler reconciles a API object
type APIReconciler struct {
	client        k8client.Client
	ods           *synchronizer.OperatorDataStore
	ch            *chan synchronizer.APIEvent
	statusUpdater *status.UpdateHandler
	mgr           manager.Manager
}

// NewAPIController creates a new API controller instance. API Controllers watches for dpv1alpha1.API and gwapiv1b1.HTTPRoute.
func NewAPIController(mgr manager.Manager, operatorDataStore *synchronizer.OperatorDataStore, statusUpdater *status.UpdateHandler,
	ch *chan synchronizer.APIEvent) error {
	r := &APIReconciler{
		client:        mgr.GetClient(),
		ods:           operatorDataStore,
		ch:            ch,
		statusUpdater: statusUpdater,
		mgr:           mgr,
	}
	c, err := controller.New(constants.APIController, mgr, controller.Options{Reconciler: r})
	if err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error creating API controller : %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2600,
		})
		return err
	}
	ctx := context.Background()

	conf := config.ReadConfigs()
	predicates := []predicate.Predicate{predicate.NewPredicateFuncs(utils.FilterByNamespaces(conf.Adapter.Operator.Namespaces))}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.API{}}, &handler.EnqueueRequestForObject{},
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching API resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2601,
		})
		return err
	}
	if err := addIndexes(ctx, mgr); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error adding indexes: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2601,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &gwapiv1b1.HTTPRoute{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIForHTTPRoute),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching HTTPRoute resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2602,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &corev1.Service{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIsForService),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching Service resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2618,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.BackendPolicy{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIsForBackendPolicy),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching BackendPolicy resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2624,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.Authentication{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIsForAuthentication),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching Authentication resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2612,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.APIPolicy{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIsForAPIPolicy),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching APIPolicy resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2612,
		})
		return err
	}

	if err := c.Watch(&source.Kind{Type: &dpv1alpha1.Scope{}}, handler.EnqueueRequestsFromMapFunc(r.getAPIsForScope),
		predicates...); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error watching scope resources: %v", err),
			Severity:  logging.BLOCKER,
			ErrorCode: 2612,
		})
		return err
	}

	loggers.LoggerAPKOperator.Info("API Controller successfully started. Watching API Objects....")
	return nil
}

//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apis/finalizers,verbs=update
//+kubebuilder:rbac:groups=dp.wso2.com,resources=httproutes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=httproutes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=httproutes/finalizers,verbs=update
//+kubebuilder:rbac:groups=dp.wso2.com,resources=authentications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=authentications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=authentications/finalizers,verbs=update
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apipolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apipolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=apipolicies/finalizers,verbs=update
//+kubebuilder:rbac:groups=dp.wso2.com,resources=scopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dp.wso2.com,resources=scopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dp.wso2.com,resources=scopes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (apiReconciler *APIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Check whether the API CR exist, if not consider as a DELETE event.
	var apiDef dpv1alpha1.API
	if err := apiReconciler.client.Get(ctx, req.NamespacedName, &apiDef); err != nil {
		apiState, found := apiReconciler.ods.GetCachedAPI(req.NamespacedName)
		if found && k8error.IsNotFound(err) {
			// The api doesn't exist in the api Cache, remove it
			apiReconciler.ods.DeleteCachedAPI(req.NamespacedName)
			loggers.LoggerAPKOperator.Infof("Delete event has received for API : %s, hence deleted from API cache", req.NamespacedName.String())
			*apiReconciler.ch <- synchronizer.APIEvent{EventType: constants.Delete, Event: apiState}
			return ctrl.Result{}, nil
		}
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message: fmt.Sprintf("Api CR related to the reconcile request with key: %s returned error."+
				" Assuming API is already deleted, hence ignoring the error : %v",
				req.NamespacedName.String(), err),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2604,
		})
		return ctrl.Result{}, nil
	}

	// Retrieve HTTPRoutes
	prodHTTPRoute, sandHTTPRoute, err := apiReconciler.resolveAPIRefs(ctx, req.Namespace, apiDef.Spec.ProdHTTPRouteRef,
		apiDef.Spec.SandHTTPRouteRef, req.NamespacedName.String())
	if err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error retrieving ref CRs for API in namespace : %s, %v", req.NamespacedName.String(), err),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2604,
		})
		return ctrl.Result{}, err
	}
	loggers.LoggerAPKOperator.Debugf("HTTPRoutes are retrieved successfully for API CR %s", req.NamespacedName.String())

	if !apiDef.Status.Accepted {
		apiState := apiReconciler.ods.AddAPIState(apiDef, prodHTTPRoute, sandHTTPRoute)
		*apiReconciler.ch <- synchronizer.APIEvent{EventType: constants.Create, Event: apiState}
		//TODO(amali) update status only after deployed without errors
		apiReconciler.handleStatus(req.NamespacedName, constants.DeployedState, []string{})
	} else if cachedAPI, events, updated :=
		apiReconciler.ods.UpdateAPIState(&apiDef, prodHTTPRoute, sandHTTPRoute); updated {
		*apiReconciler.ch <- synchronizer.APIEvent{EventType: constants.Update, Event: cachedAPI}
		apiReconciler.handleStatus(req.NamespacedName, constants.UpdatedState, events)
	}
	return ctrl.Result{}, nil
}

// resolveAPIRefs validates following references related to the API
// - HTTPRoutes
func (apiReconciler *APIReconciler) resolveAPIRefs(ctx context.Context, namespace, prodHTTPRouteRef, sandHTTPRouteRef,
	api string) (*synchronizer.HTTPRouteState, *synchronizer.HTTPRouteState, error) {
	prodHTTPRoute := &synchronizer.HTTPRouteState{
		HTTPRoute: &gwapiv1b1.HTTPRoute{},
	}
	sandHTTPRoute := &synchronizer.HTTPRouteState{
		HTTPRoute: &gwapiv1b1.HTTPRoute{},
	}
	var err error
	if prodHTTPRouteRef != "" {
		if prodHTTPRoute, err = apiReconciler.resolveHTTPRouteRefs(ctx, namespace, prodHTTPRouteRef, api); err != nil {
			return nil, nil, fmt.Errorf("error while resolving production httpRouteref %s in namespace :%s has not found. %s",
				prodHTTPRouteRef, namespace, err.Error())
		}
	}

	if sandHTTPRouteRef != "" {
		if sandHTTPRoute, err = apiReconciler.resolveHTTPRouteRefs(ctx, namespace, sandHTTPRouteRef, api); err != nil {
			return nil, nil, fmt.Errorf("error while resolving sandbox httpRouteref %s in namespace :%s has not found. %s",
				sandHTTPRouteRef, namespace, err.Error())
		}
	}
	return prodHTTPRoute, sandHTTPRoute, nil
}

// resolveHTTPRouteRefs validates following references related to the API
// - Authentications
func (apiReconciler *APIReconciler) resolveHTTPRouteRefs(ctx context.Context, namespace, httpRouteRef, api string) (*synchronizer.HTTPRouteState, error) {
	httpRouteState := &synchronizer.HTTPRouteState{
		HTTPRoute: &gwapiv1b1.HTTPRoute{},
	}
	var err error
	if err := apiReconciler.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: httpRouteRef},
		httpRouteState.HTTPRoute); err != nil {
		return nil, fmt.Errorf("error while getting httproute %s in namespace :%s, %s", httpRouteRef, namespace, err.Error())
	}
	if httpRouteState.Authentications, err = apiReconciler.getAuthenticationsForHTTPRoute(ctx, httpRouteState.HTTPRoute); err != nil {
		return nil, fmt.Errorf("error while getting httproute auth defaults %s in namespace :%s, %s", httpRouteRef,
			namespace, err.Error())
	}
	if httpRouteState.ResourceAuthentications, err = apiReconciler.getAuthenticationsForResources(ctx, httpRouteState.HTTPRoute); err != nil {
		return nil, fmt.Errorf("error while getting httproute auth defaults %s in namespace :%s, %s", httpRouteRef,
			namespace, err.Error())
	}
	if httpRouteState.APIPolicies, err = apiReconciler.getAPIPoliciesForHTTPRoute(ctx, httpRouteState.HTTPRoute); err != nil {
		return nil, fmt.Errorf("error while getting httproute auth defaults %s in namespace :%s, %s", httpRouteRef,
			namespace, err.Error())
	}
	if httpRouteState.ResourceAPIPolicies, err = apiReconciler.getAPIPoliciesForResources(ctx, httpRouteState.HTTPRoute); err != nil {
		return nil, fmt.Errorf("error while getting httproute auth defaults %s in namespace :%s, %s", httpRouteRef,
			namespace, err.Error())
	}
	httpRouteState.BackendPropertyMapping = apiReconciler.getBackendProperties(ctx, httpRouteState.HTTPRoute)
	httpRouteState.Scopes, err = apiReconciler.getScopesForHTTPRoute(ctx, httpRouteState.HTTPRoute, api)

	return httpRouteState, err
}

func (apiReconciler *APIReconciler) getAuthenticationsForHTTPRoute(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute) (map[string]dpv1alpha1.Authentication, error) {
	authentications := make(map[string]dpv1alpha1.Authentication)
	authenticationList := &dpv1alpha1.AuthenticationList{}
	if err := apiReconciler.client.List(ctx, authenticationList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAuthenticationIndex, utils.NamespacedName(httpRoute).String()),
	}); err != nil {
		return nil, err
	}
	for _, item := range authenticationList.Items {
		authentications[utils.NamespacedName(&item).String()] = item
	}
	return authentications, nil
}

func (apiReconciler *APIReconciler) getScopesForHTTPRoute(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute, api string) (map[string]dpv1alpha1.Scope, error) {
	scopes := make(map[string]dpv1alpha1.Scope)
	for _, rule := range httpRoute.Spec.Rules {
		for _, filter := range rule.Filters {
			if filter.Type == gwapiv1b1.HTTPRouteFilterExtensionRef && filter.ExtensionRef != nil && filter.ExtensionRef.Kind == constants.KindScope {
				scope := &dpv1alpha1.Scope{}
				if err := apiReconciler.client.Get(ctx, types.NamespacedName{Namespace: httpRoute.Namespace, Name: string(filter.ExtensionRef.Name)},
					scope); err != nil {
					return nil, fmt.Errorf("error while getting scope %s in namespace :%s, %s", filter.ExtensionRef.Name,
						httpRoute.Namespace, err.Error())
				}
				scopes[utils.NamespacedName(scope).String()] = *scope
				setIndexScopeForAPI(ctx, apiReconciler.mgr, scope, api)
			}
		}
	}

	return scopes, nil
}

func (apiReconciler *APIReconciler) getAuthenticationsForResources(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute) (map[string]dpv1alpha1.Authentication, error) {
	authentications := make(map[string]dpv1alpha1.Authentication)
	authenticationList := &dpv1alpha1.AuthenticationList{}
	if err := apiReconciler.client.List(ctx, authenticationList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAuthenticationResourceIndex, utils.NamespacedName(httpRoute).String()),
	}); err != nil {
		return nil, err
	}
	for _, item := range authenticationList.Items {
		authentications[utils.NamespacedName(&item).String()] = item
	}
	return authentications, nil
}

func (apiReconciler *APIReconciler) getAPIPoliciesForHTTPRoute(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute) (map[string]dpv1alpha1.APIPolicy, error) {
	apiPolicies := make(map[string]dpv1alpha1.APIPolicy)
	apiPolicyList := &dpv1alpha1.APIPolicyList{}
	if err := apiReconciler.client.List(ctx, apiPolicyList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIPolicyIndex, utils.NamespacedName(httpRoute).String()),
	}); err != nil {
		return nil, err
	}
	for _, item := range apiPolicyList.Items {
		apiPolicies[utils.NamespacedName(&item).String()] = item
	}
	return apiPolicies, nil
}

func (apiReconciler *APIReconciler) getAPIPoliciesForResources(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute) (map[string]dpv1alpha1.APIPolicy, error) {
	apiPolicies := make(map[string]dpv1alpha1.APIPolicy)
	apiPolicyList := &dpv1alpha1.APIPolicyList{}
	if err := apiReconciler.client.List(ctx, apiPolicyList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIPolicyResourceIndex, utils.NamespacedName(httpRoute).String()),
	}); err != nil {
		return nil, err
	}
	for _, item := range apiPolicyList.Items {
		apiPolicies[utils.NamespacedName(&item).String()] = item
	}
	return apiPolicies, nil
}

func (apiReconciler *APIReconciler) getBackendProperties(ctx context.Context,
	httpRoute *gwapiv1b1.HTTPRoute) dpv1alpha1.BackendPropertyMapping {
	backendPropertyMapping := make(dpv1alpha1.BackendPropertyMapping)
	for _, rule := range httpRoute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			backendNamespacedName := types.NamespacedName{
				Name:      string(backend.Name),
				Namespace: utils.GetNamespace(backend.Namespace, httpRoute.Namespace),
			}
			tls, protocol := apiReconciler.getBackendConfigs(ctx, backendNamespacedName)
			backendPropertyMapping[backendNamespacedName] = dpv1alpha1.BackendProperties{
				ResolvedHostname: apiReconciler.getHostNameForBackend(ctx,
					backend, httpRoute.Namespace),
				TLS:      tls,
				Protocol: protocol,
			}
		}
	}
	loggers.LoggerAPKOperator.Debugf("Generated backendPropertyMapping: %v", backendPropertyMapping)
	return backendPropertyMapping
}

// getHostNameForService resolves the backed hostname for services.
// When service type is ExternalName then ExternalName property is used as the hostname.
// Otherwise defaulted to service name as <namespace>.<service>
func (apiReconciler *APIReconciler) getHostNameForBackend(ctx context.Context, backend gwapiv1b1.HTTPBackendRef,
	defaultNamespace string) string {
	var service = new(corev1.Service)
	err := apiReconciler.client.Get(context.Background(), types.NamespacedName{
		Name:      string(backend.Name),
		Namespace: utils.GetNamespace(backend.Namespace, defaultNamespace)}, service)
	if err == nil {
		switch service.Spec.Type {
		case corev1.ServiceTypeExternalName:
			return service.Spec.ExternalName
		}
	}
	return utils.GetDefaultHostNameForBackend(backend, defaultNamespace)
}

// getTLSConfigForBackend resolves backend TLS configurations.
func (apiReconciler *APIReconciler) getBackendConfigs(ctx context.Context,
	serviceNamespacedName types.NamespacedName) (dpv1alpha1.TLSConfig, dpv1alpha1.BackendProtocolType) {
	tlsConfig := dpv1alpha1.TLSConfig{}
	protocol := dpv1alpha1.HTTPProtocol
	backendPolicyList := &dpv1alpha1.BackendPolicyList{}
	if err := apiReconciler.client.List(ctx, backendPolicyList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceBackendPolicyIndex, serviceNamespacedName.String()),
	}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated BackendPolicies for service: %s", serviceNamespacedName),
			Severity:  logging.CRITICAL,
			ErrorCode: 2611,
		})
	}
	if len(backendPolicyList.Items) > 0 {
		backendPolicy := *utils.TieBreaker(utils.GetPtrSlice(backendPolicyList.Items))
		var backendProtocol dpv1alpha1.BackendProtocolType
		if backendPolicy.Spec.Override != nil {
			tlsConfig = backendPolicy.Spec.Override.TLS
			backendProtocol = backendPolicy.Spec.Override.Protocol
		} else if backendPolicy.Spec.Default != nil {
			tlsConfig = backendPolicy.Spec.Default.TLS
			backendProtocol = backendPolicy.Spec.Default.Protocol
		}
		if len(backendProtocol) > 0 {
			switch protocol {
			case dpv1alpha1.HTTPProtocol:
				fallthrough
			case dpv1alpha1.HTTPSProtocol:
				fallthrough
			case dpv1alpha1.WSProtocol:
				fallthrough
			case dpv1alpha1.WSSProtocol:
				protocol = backendProtocol
			default:
				protocol = dpv1alpha1.HTTPProtocol
			}
		}
	}
	return tlsConfig, protocol
}

// getAPIForHTTPRoute triggers the API controller reconcile method based on the changes detected
// from HTTPRoute objects. If the changes are done for an API stored in the Operator Data store,
// a new reconcile event will be created and added to the reconcile event queue.
func (apiReconciler *APIReconciler) getAPIForHTTPRoute(obj k8client.Object) []reconcile.Request {
	ctx := context.Background()
	httpRoute, ok := obj.(*gwapiv1b1.HTTPRoute)
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unexpected object type, bypassing reconciliation: %v", httpRoute),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2608,
		})
		return []reconcile.Request{}
	}

	apiList := &dpv1alpha1.APIList{}
	if err := apiReconciler.client.List(ctx, apiList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIIndex, utils.NamespacedName(httpRoute).String()),
	}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated APIs: %s", utils.NamespacedName(httpRoute).String()),
			Severity:  logging.CRITICAL,
			ErrorCode: 2610,
		})
		return []reconcile.Request{}
	}

	if len(apiList.Items) == 0 {
		loggers.LoggerAPKOperator.Debugf("APIs for HTTPRoute not found: %s", utils.NamespacedName(httpRoute).String())
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}
	for _, api := range apiList.Items {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      api.Name,
				Namespace: api.Namespace},
		}
		requests = append(requests, req)
		loggers.LoggerAPKOperator.Infof("Adding reconcile request for API: %s/%s", api.Namespace, api.Name)
	}
	return requests
}

// getAPIsForService triggers the API controller reconcile method based on the changes detected
// from Service objects. This generates a reconcile request for a API looking up two indexes;
// serviceHTTPRouteIndex and httpRouteAPIIndex in that order.
func (apiReconciler *APIReconciler) getAPIsForService(obj k8client.Object) []reconcile.Request {
	ctx := context.Background()
	service, ok := obj.(*corev1.Service)
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unexpected object type, bypassing reconciliation: %v", service),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2620,
		})
		return []reconcile.Request{}
	}

	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := apiReconciler.client.List(ctx, httpRouteList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(serviceHTTPRouteIndex, utils.NamespacedName(service).String()),
	}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated HTTPRoutes: %s", utils.NamespacedName(service).String()),
			Severity:  logging.CRITICAL,
			ErrorCode: 2621,
		})
		return []reconcile.Request{}
	}

	if len(httpRouteList.Items) == 0 {
		loggers.LoggerAPKOperator.Debugf("HTTPRoutes for Service not found: %s", utils.NamespacedName(service).String())
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}
	for _, httpRoute := range httpRouteList.Items {
		requests = append(requests, apiReconciler.getAPIForHTTPRoute(&httpRoute)...)
	}
	return requests
}

// getAPIsForBackendPolicy triggers the API controller reconcile method based on the changes detected
// from BackendPolicy objects using the targetRef to a Service object.
func (apiReconciler *APIReconciler) getAPIsForBackendPolicy(obj k8client.Object) []reconcile.Request {
	ctx := context.Background()
	backendPolicy, ok := obj.(*dpv1alpha1.BackendPolicy)
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unexpected object type, bypassing reconciliation: %v", backendPolicy),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2622,
		})
		return []reconcile.Request{}
	}

	service := &corev1.Service{}
	if err := apiReconciler.client.Get(ctx, types.NamespacedName{
		Name: string(backendPolicy.Spec.TargetRef.Name),
		Namespace: utils.GetNamespace((*gwapiv1b1.Namespace)(backendPolicy.Spec.TargetRef.Namespace),
			backendPolicy.Namespace),
	}, service); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated Service for BackendPolicy: %s", utils.NamespacedName(backendPolicy).String()),
			Severity:  logging.CRITICAL,
			ErrorCode: 2623,
		})
		return []reconcile.Request{}
	}
	return apiReconciler.getAPIsForService(service)
}

// getAPIForAuthentication triggers the API controller reconcile method based on the changes detected
// from Authentication objects. If the changes are done for an API stored in the Operator Data store,
// a new reconcile event will be created and added to the reconcile event queue.
func (apiReconciler *APIReconciler) getAPIsForAuthentication(obj k8client.Object) []reconcile.Request {
	authentication, ok := obj.(*dpv1alpha1.Authentication)
	ctx := context.Background()
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("unexpected object type, bypassing reconciliation: %v", authentication),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2608,
		})
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}

	// todo(amali) move this validation to validation hook
	if !(authentication.Spec.TargetRef.Kind == constants.KindHTTPRoute || authentication.Spec.TargetRef.Kind == constants.KindResource) {
		loggers.LoggerAPKOperator.Errorf("Unsupported target ref kind : %s was given for authentication: %s",
			authentication.Spec.TargetRef.Kind, authentication.Name)
		return requests
	}
	loggers.LoggerAPKOperator.Debugf("Finding reconcile API requests for httpRoute: %s in namespace : %s",
		authentication.Spec.TargetRef.Name, authentication.Namespace)

	apiList := &dpv1alpha1.APIList{}

	namespacedName := types.NamespacedName{
		Name: string(authentication.Spec.TargetRef.Name),
		Namespace: utils.GetNamespace(
			(*gwapiv1b1.Namespace)(authentication.Spec.TargetRef.Namespace),
			authentication.Namespace)}.String()

	if err := apiReconciler.client.List(ctx, apiList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIIndex, namespacedName)}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated APIs: %s", namespacedName),
			Severity:  logging.CRITICAL,
			ErrorCode: 2610,
		})
		return []reconcile.Request{}
	}

	if len(apiList.Items) == 0 {
		loggers.LoggerAPKOperator.Debugf("APIs for HTTPRoute not found: %s", namespacedName)
		return []reconcile.Request{}
	}

	for _, api := range apiList.Items {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      api.Name,
				Namespace: api.Namespace},
		}
		requests = append(requests, req)
		loggers.LoggerAPKOperator.Infof("Adding reconcile request for API: %s/%s", api.Namespace, api.Name)
	}
	return requests
}

// getAPIForAuthentication triggers the API controller reconcile method based on the changes detected
// from Authentication objects. If the changes are done for an API stored in the Operator Data store,
// a new reconcile event will be created and added to the reconcile event queue.
func (apiReconciler *APIReconciler) getAPIsForAPIPolicy(obj k8client.Object) []reconcile.Request {
	apiPolicy, ok := obj.(*dpv1alpha1.APIPolicy)
	ctx := context.Background()
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("unexpected object type, bypassing reconciliation: %v", apiPolicy),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2608,
		})
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}

	// todo(amali) move this validation to validation hook
	if !(apiPolicy.Spec.TargetRef.Kind == constants.KindHTTPRoute || apiPolicy.Spec.TargetRef.Kind == constants.KindResource) {
		loggers.LoggerAPKOperator.Errorf("Unsupported target ref kind : %s was given for authentication: %s",
			apiPolicy.Spec.TargetRef.Kind, apiPolicy.Name)
		return requests
	}
	loggers.LoggerAPKOperator.Debugf("Finding reconcile API requests for httpRoute: %s in namespace : %s",
		apiPolicy.Spec.TargetRef.Name, apiPolicy.Namespace)

	apiList := &dpv1alpha1.APIList{}

	namespacedName := types.NamespacedName{
		Name: string(apiPolicy.Spec.TargetRef.Name),
		Namespace: utils.GetNamespace(
			(*gwapiv1b1.Namespace)(apiPolicy.Spec.TargetRef.Namespace),
			apiPolicy.Namespace)}.String()

	if err := apiReconciler.client.List(ctx, apiList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIIndex, namespacedName)}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated APIs: %s", namespacedName),
			Severity:  logging.CRITICAL,
			ErrorCode: 2610,
		})
		return []reconcile.Request{}
	}

	if len(apiList.Items) == 0 {
		loggers.LoggerAPKOperator.Debugf("APIs for HTTPRoute not found: %s", namespacedName)
		return []reconcile.Request{}
	}

	for _, api := range apiList.Items {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      api.Name,
				Namespace: api.Namespace},
		}
		requests = append(requests, req)
		loggers.LoggerAPKOperator.Infof("Adding reconcile request for API: %s/%s", api.Namespace, api.Name)
	}
	return requests
}

// getAPIsForScope triggers the API controller reconcile method based on the changes detected
// from scope objects. If the changes are done for an API stored in the Operator Data store,
// a new reconcile event will be created and added to the reconcile event queue.
func (apiReconciler *APIReconciler) getAPIsForScope(obj k8client.Object) []reconcile.Request {
	scope, ok := obj.(*dpv1alpha1.Scope)
	ctx := context.Background()
	if !ok {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("unexpected object type, bypassing reconciliation: %v", scope),
			Severity:  logging.TRIVIAL,
			ErrorCode: 2608,
		})
		return []reconcile.Request{}
	}

	requests := []reconcile.Request{}

	apiList := &dpv1alpha1.APIList{}

	if err := apiReconciler.client.List(ctx, apiList, &k8client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(httpRouteAPIIndex, utils.NamespacedName(scope).String())}); err != nil {
		loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Unable to find associated APIs: %s", utils.NamespacedName(scope).String()),
			Severity:  logging.CRITICAL,
			ErrorCode: 2610,
		})
		return []reconcile.Request{}
	}

	if len(apiList.Items) == 0 {
		loggers.LoggerAPKOperator.Debugf("APIs for HTTPRoute not found: %s", utils.NamespacedName(scope).String())
		return []reconcile.Request{}
	}

	for _, api := range apiList.Items {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      api.Name,
				Namespace: api.Namespace},
		}
		requests = append(requests, req)
		loggers.LoggerAPKOperator.Infof("Adding reconcile request for API: %s/%s", api.Namespace, api.Name)
	}
	return requests
}

// addIndexes adds indexing on API, for
//   - prodution and sandbox HTTPRoutes
//     referenced in API objects via `.spec.prodHTTPRouteRef` and `.spec.sandHTTPRouteRef`
//     This helps to find APIs that are affected by a HTTPRoute CRUD operation.
//   - authentications
//     authentication schemes related to httproutes
//     This helps to find authentication schemes binded to HTTPRoute.
//   - apiPolicies
//     apiPolicy schemes related to httproutes
//     This helps to find apiPolicy schemes binded to HTTPRoute.
func addIndexes(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.API{}, httpRouteAPIIndex,
		func(rawObj k8client.Object) []string {
			api := rawObj.(*dpv1alpha1.API)
			var httpRoutes []string
			if api.Spec.ProdHTTPRouteRef != "" {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: api.Namespace,
						Name:      api.Spec.ProdHTTPRouteRef,
					}.String())
			}
			if api.Spec.SandHTTPRouteRef != "" {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: api.Namespace,
						Name:      api.Spec.SandHTTPRouteRef,
					}.String())
			}
			return httpRoutes
		}); err != nil {
		return err
	}

	// Service (BackendRefs) to HTTPRoute indexer
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.HTTPRoute{}, serviceHTTPRouteIndex,
		func(rawObj k8client.Object) []string {
			httpRoute := rawObj.(*gwapiv1b1.HTTPRoute)
			var services []string
			for _, rule := range httpRoute.Spec.Rules {
				for _, backendRef := range rule.BackendRefs {
					services = append(services, types.NamespacedName{
						Namespace: utils.GetNamespace(backendRef.Namespace,
							httpRoute.ObjectMeta.Namespace),
						Name: string(backendRef.Name),
					}.String())
				}
			}
			return services
		}); err != nil {
		return err
	}

	// Service to BackendPolicy indexer
	if err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.BackendPolicy{}, serviceBackendPolicyIndex,
		func(rawObj k8client.Object) []string {
			backendPolicy := rawObj.(*dpv1alpha1.BackendPolicy)
			var services []string
			if backendPolicy.Spec.TargetRef.Kind == constants.KindService {
				services = append(services,
					types.NamespacedName{
						Name:      string(backendPolicy.Spec.TargetRef.Name),
						Namespace: backendPolicy.Namespace,
					}.String())
			}
			return services
		}); err != nil {
		return err
	}

	// authentication to HTTPRoute indexer
	if err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.Authentication{}, httpRouteAuthenticationIndex,
		func(rawObj k8client.Object) []string {
			authentication := rawObj.(*dpv1alpha1.Authentication)
			var httpRoutes []string
			if authentication.Spec.TargetRef.Kind == constants.KindHTTPRoute {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: utils.GetNamespace(
							(*gwapiv1b1.Namespace)(authentication.Spec.TargetRef.Namespace),
							authentication.Namespace),
						Name: string(authentication.Spec.TargetRef.Name),
					}.String())
			}
			return httpRoutes
		}); err != nil {
		return err
	}

	// Till the below is httproute rule name and targetref sectionname is supported,
	// https://gateway-api.sigs.k8s.io/geps/gep-713/?h=multiple+targetrefs#apply-policies-to-sections-of-a-resource-future-extension
	// we will use a temporary kindName called Resource for policy attachments
	// TODO(amali) Fix after the official support is available
	if err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.Authentication{}, httpRouteAuthenticationResourceIndex,
		func(rawObj k8client.Object) []string {
			authentication := rawObj.(*dpv1alpha1.Authentication)
			var httpRoutes []string
			if authentication.Spec.TargetRef.Kind == constants.KindResource {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: utils.GetNamespace(
							(*gwapiv1b1.Namespace)(authentication.Spec.TargetRef.Namespace),
							authentication.Namespace),
						Name: string(authentication.Spec.TargetRef.Name),
					}.String())
			}
			return httpRoutes
		}); err != nil {
		return err
	}

	// APIPolicy to HTTPRoute indexer
	if err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.APIPolicy{}, httpRouteAPIPolicyIndex,
		func(rawObj k8client.Object) []string {
			apiPolicy := rawObj.(*dpv1alpha1.APIPolicy)
			var httpRoutes []string
			if apiPolicy.Spec.TargetRef.Kind == constants.KindHTTPRoute {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: utils.GetNamespace(
							(*gwapiv1b1.Namespace)(apiPolicy.Spec.TargetRef.Namespace), apiPolicy.Namespace),
						Name: string(apiPolicy.Spec.TargetRef.Name),
					}.String())
			}
			return httpRoutes
		}); err != nil {
		return err
	}

	// Till the below is httproute rule name and targetref sectionname is supported,
	// https://gateway-api.sigs.k8s.io/geps/gep-713/?h=multiple+targetrefs#apply-policies-to-sections-of-a-resource-future-extension
	// we will use a temporary kindName called Resource for policy attachments
	// TODO(amali) Fix after the official support is available
	err := mgr.GetFieldIndexer().IndexField(ctx, &dpv1alpha1.APIPolicy{}, httpRouteAPIPolicyResourceIndex,
		func(rawObj k8client.Object) []string {
			apiPolicy := rawObj.(*dpv1alpha1.APIPolicy)
			var httpRoutes []string
			if apiPolicy.Spec.TargetRef.Kind == constants.KindResource {
				httpRoutes = append(httpRoutes,
					types.NamespacedName{
						Namespace: utils.GetNamespace(
							(*gwapiv1b1.Namespace)(apiPolicy.Spec.TargetRef.Namespace),
							apiPolicy.Namespace),
						Name: string(apiPolicy.Spec.TargetRef.Name),
					}.String())
			}
			return httpRoutes
		})
	return err
}

// setIndexScopeForAPI sets the index for the Scope CR to API CR
func setIndexScopeForAPI(ctx context.Context, mgr manager.Manager, scope *dpv1alpha1.Scope, api string) error {
	return mgr.GetFieldIndexer().IndexField(ctx, scope, apiScopeIndex,
		func(rawObj k8client.Object) []string {
			return []string{api}
		})
}

// handleStatus updates the API CR update
func (apiReconciler *APIReconciler) handleStatus(apiKey types.NamespacedName, state string, events []string) {
	accept := false
	message := ""
	event := ""

	switch state {
	case constants.DeployedState:
		accept = true
		message = "API is deployed to the gateway."
	case constants.UpdatedState:
		accept = true
		message = fmt.Sprintf("API update is deployed to the gateway. %v Updated", events)
	}
	timeNow := metav1.Now()
	event = fmt.Sprintf("[%s] %s", timeNow.String(), message)

	apiReconciler.statusUpdater.Send(status.Update{
		NamespacedName: apiKey,
		Resource:       new(dpv1alpha1.API),
		UpdateStatus: func(obj k8client.Object) k8client.Object {
			h, ok := obj.(*dpv1alpha1.API)
			if !ok {
				loggers.LoggerAPKOperator.ErrorC(logging.ErrorDetails{
					Message:   fmt.Sprintf("Unsupported object type %T", obj),
					Severity:  logging.BLOCKER,
					ErrorCode: 2617,
				})
			}
			hCopy := h.DeepCopy()
			hCopy.Status.Status = state
			hCopy.Status.Accepted = accept
			hCopy.Status.Message = message
			hCopy.Status.Events = append(hCopy.Status.Events, event)
			hCopy.Status.TransitionTime = &timeNow
			return hCopy
		},
	})
}
