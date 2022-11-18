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

package database

import (
	"fmt"
	"time"

	apkmgt "github.com/wso2/apk/adapter/pkg/discovery/api/wso2/discovery/apkmgt"
	apiProtos "github.com/wso2/apk/adapter/pkg/discovery/api/wso2/discovery/service/apkmgt"
	"github.com/wso2/apk/adapter/pkg/logging"
	"github.com/wso2/apk/management-server/internal/logger"
)

var DbCache *ApplicationLocalCache

func init() {
	DbCache = NewApplicationLocalCache(cleanupInterval)
}

func GetApplicationByUUID(uuid string) (*apkmgt.Application, error) {
	rows, _ := ExecDBQuery(QueryGetApplicationByUUID, uuid)
	rows.Next()
	values, err := rows.Values()
	if err != nil {
		return nil, err
	} else {
		subs, _ := getSubscriptionsForApplication(uuid)
		keys, _ := getConsumerKeysForApplication(uuid)
		application := &apkmgt.Application{
			Uuid:          values[0].(string),
			Name:          values[1].(string),
			Owner:         "",  //ToDo : Check how to get Owner from db
			Attributes:    nil, //ToDo : check the values for Attributes
			Subscriber:    "",
			Organization:  values[3].(string),
			Subscriptions: subs,
			ConsumerKeys:  keys,
		}
		DbCache.Update(application, time.Now().Unix()+ttl.Microseconds())
		return application, nil
	}
}

// GetCachedApplicationByUUID returns the Application details from the cache.
// If the application is not available in the cache, it will fetch the application from DB.
func GetCachedApplicationByUUID(uuid string) (*apkmgt.Application, error) {
	if app, ok := DbCache.Read(uuid); ok == nil {
		return &app, nil
	} else {
		return GetApplicationByUUID(uuid)
	}
}

// getSubscriptionsForApplication returns all subscriptions from DB, for a given application.
func getSubscriptionsForApplication(appUuid string) ([]*apkmgt.Subscription, error) {
	rows, err := ExecDBQuery(QueryGetAllSubscriptionsForApplication, appUuid)
	if err != nil {
	}
	var subs []*apkmgt.Subscription
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		} else {
			subs = append(subs, &apkmgt.Subscription{
				Uuid:               values[0].(string),
				ApiUuid:            values[1].(string),
				PolicyId:           "",
				SubscriptionStatus: values[3].(string),
				Organization:       values[4].(string),
				CreatedBy:          values[5].(string),
			})
		}
	}
	return subs, nil
}

// getConsumerKeysForApplication returns all Consumer Keys from DB, for a given application.
func getConsumerKeysForApplication(appUUID string) ([]*apkmgt.ConsumerKey, error) {
	rows, err := ExecDBQuery(QueryConsumerKeysForApplication, appUUID)
	if err != nil {
	}
	var keys []*apkmgt.ConsumerKey
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		} else {
			keys = append(keys, &apkmgt.ConsumerKey{
				Key:        values[0].(string),
				KeyManager: values[1].(string),
			})
		}
	}
	return keys, nil
}

// GetSubscriptionByUUID returns the Application details from the DB for a given subscription UUID.
func GetSubscriptionByUUID(subUUID string) (*apkmgt.Subscription, error) {
	rows, _ := ExecDBQuery(QuerySubscriptionByUUID, subUUID)
	rows.Next()
	values, err := rows.Values()
	if err != nil {
		return nil, err
	} else {
		return &apkmgt.Subscription{
			Uuid:               values[0].(string),
			ApiUuid:            values[1].(string),
			PolicyId:           "",
			SubscriptionStatus: values[2].(string),
			Organization:       values[3].(string),
			CreatedBy:          values[4].(string),
		}, nil
	}
}

func CreateAPI(api *apiProtos.API) error {
	_, err := ExecDBQuery(QueryCreateAPI, &api.Uuid, &api.Name, &api.Provider,
		&api.Version, &api.Context, &api.OrganizationId, &api.CreatedBy, time.Now(), &api.Type)

	if err != nil {
		logger.LoggerDatabase.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error creating API %q, Error: %v", api.Uuid, err.Error()),
			Severity:  logging.CRITICAL,
			ErrorCode: 1201,
		})
		return err
	}
	return nil
}

func UpdateAPI(api *apiProtos.API) error {
	_, err := ExecDBQuery(QueryUpdateAPI, &api.Uuid, &api.Name, &api.Provider,
		&api.Version, &api.Context, &api.OrganizationId, &api.UpdatedBy, time.Now(), &api.Type)
	if err != nil {
		logger.LoggerDatabase.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error updating API %q, Error: %v", api.Uuid, err.Error()),
			Severity:  logging.CRITICAL,
			ErrorCode: 1202,
		})
		return err
	}
	return nil
}

func DeleteAPI(api *apiProtos.API) error {
	_, err := ExecDBQuery(QueryDeleteAPI, api.Uuid)
	if err != nil {
		logger.LoggerDatabase.ErrorC(logging.ErrorDetails{
			Message:   fmt.Sprintf("Error deleting API %q, Error: %v", api.Uuid, err.Error()),
			Severity:  logging.CRITICAL,
			ErrorCode: 1203,
		})
		return err
	}
	return nil
}
