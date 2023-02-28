//
// Copyright (c) 2022, WSO2 LLC. (http://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//

import ballerina/http;
import ballerina/log;

service /api/am/admin on ep0 {
    
    isolated resource function get 'application\-rate\-plans(@http:Header string? accept = "application/json") returns ApplicationRatePlanList|NotAcceptableError|BadRequestError|InternalServerErrorError {
        ApplicationRatePlanList|APKError appPolicyList = getApplicationUsagePlans();
        if appPolicyList is ApplicationRatePlanList {
            log:printDebug(appPolicyList.toString());
            return appPolicyList;
        } else {
            return handleAPKError(appPolicyList);
        }
    }
    isolated resource function post 'application\-rate\-plans(@http:Payload ApplicationRatePlan payload, @http:Header string 'content\-type = "application/json") returns CreatedApplicationRatePlan|BadRequestError|UnsupportedMediaTypeError|InternalServerErrorError|error {
        ApplicationRatePlan|APKError createdAppPol = addApplicationUsagePlan(payload);
        if createdAppPol is ApplicationRatePlan {
            log:printDebug(createdAppPol.toString());
            CreatedApplicationRatePlan crPol = {body: check createdAppPol.cloneWithType(ApplicationRatePlan)};
            return crPol;
        } else {
            return handleAPKError(createdAppPol);
        }
    }
    isolated resource function get 'application\-rate\-plans/[string planId]() returns ApplicationRatePlan|NotFoundError|NotAcceptableError|BadRequestError|InternalServerErrorError {
        ApplicationRatePlan|APKError|NotFoundError appPolicy = getApplicationUsagePlanById(planId);
        if appPolicy is ApplicationRatePlan|NotFoundError {
            log:printDebug(appPolicy.toString());
            return appPolicy;
        } else {
            return handleAPKError(appPolicy);
        }
    }
    isolated resource function put 'application\-rate\-plans/[string planId](@http:Payload ApplicationRatePlan payload, @http:Header string 'content\-type = "application/json") returns ApplicationRatePlan|BadRequestError|NotFoundError|InternalServerErrorError {
        ApplicationRatePlan|NotFoundError|APKError appPolicy = updateApplicationUsagePlan(planId, payload);
        if appPolicy is ApplicationRatePlan|NotFoundError {
            log:printDebug(appPolicy.toString());
            return appPolicy;
        } else {
            return handleAPKError(appPolicy);
        }
    }
    isolated resource function delete 'application\-rate\-plans/[string planId]() returns http:Ok|NotFoundError|BadRequestError|InternalServerErrorError {
        string|APKError ex = removeApplicationUsagePlan(planId);
        if ex is APKError {
            return handleAPKError(ex);
        } else {
            return http:OK;
        }
    }
    isolated resource function get 'deny\-policies(@http:Header string? accept = "application/json") returns BlockingConditionList|NotAcceptableError|BadRequestError|InternalServerErrorError {
        BlockingConditionList|APKError conditionList = getAllDenyPolicies();
        if conditionList is BlockingConditionList {
            return conditionList;
        } else {
            return handleAPKError(conditionList);
        }
    }
    isolated resource function get 'deny\-policies/[string policyId]() returns BlockingCondition|NotFoundError|BadRequestError|NotAcceptableError|InternalServerErrorError {
        BlockingCondition|APKError|NotFoundError denyPolicy = getDenyPolicyById(policyId);
        if denyPolicy is BlockingCondition|NotFoundError {
            log:printDebug(denyPolicy.toString());
            return denyPolicy;
        } else {
            return handleAPKError(denyPolicy);
        }
    }
    isolated resource function delete 'deny\-policies/[string policyId]() returns http:Ok|NotFoundError|BadRequestError|InternalServerErrorError {
        string|APKError ex = removeDenyPolicy(policyId);
        if ex is APKError {
            return handleAPKError(ex);
        } else {
            return http:OK;
        }
    }
    isolated resource function put 'api\-categories/[string apiCategoryId](@http:Payload APICategory payload) returns APICategory|BadRequestError|NotFoundError|InternalServerErrorError|error {
        APICategory|NotFoundError|APKError  apiCategory = updateAPICategory(apiCategoryId, payload);
        if apiCategory is APICategory | NotFoundError {
            return apiCategory;
        } else {
            return handleAPKError(apiCategory);
        }
    }
    isolated resource function delete 'api\-categories/[string apiCategoryId]() returns http:Ok|NotFoundError|BadRequestError|InternalServerErrorError|error {
        string|APKError ex = removeAPICategory(apiCategoryId);
        if ex is APKError {
            return handleAPKError(ex);
        } else {
            return http:OK;
        }
    }
}

isolated function handleAPKError(APKError errorDetail) returns InternalServerErrorError|BadRequestError {
    ErrorHandler & readonly detail = errorDetail.detail();
    if detail.statusCode=="400" {
        BadRequestError badRequest = {body: {code: detail.code, message: detail.message}};
        return badRequest;
    }
    InternalServerErrorError internalServerError = {body: {code: detail.code, message: detail.message}};
    return internalServerError;
}
