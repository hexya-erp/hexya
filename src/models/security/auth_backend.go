// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package security

import (
	"fmt"

	"github.com/hexya-erp/hexya/src/models/types"
)

// AuthenticationRegistry is the authentication registry of the application
var AuthenticationRegistry *AuthBackendRegistry

// A UserNotFoundError should be returned by backends when the user is not known
type UserNotFoundError string

// Error returns the error message
func (unfe UserNotFoundError) Error() string {
	return fmt.Sprintf("User not found %s", string(unfe))
}

// A InvalidCredentialsError should be returned by backends when the user is known
// to this backend but cannot be authenticated.
type InvalidCredentialsError string

// Error returns the error message
func (ice InvalidCredentialsError) Error() string {
	return fmt.Sprintf("Wrong credentials for user %s", string(ice))
}

// An AuthBackend is an interface that is capable of authenticating a
// user and tell whether a user is a member of a given group.
type AuthBackend interface {
	// Authenticate the user defined by login and secret. Additional data
	// needed by the authentication backend may be passed into the context.
	//
	// On success, it returns the ID of the authenticated user.
	// On failure, it should return a UserNotFoundError if this user is not
	// known to this backend or a InvalidCredentialsError if it is known but
	// cannot be authenticated.
	Authenticate(login, secret string, context *types.Context) (int64, error)
}

// An AuthBackendRegistry holds an ordered list of AuthBackend instances
// that enables authentication against several backends.
// A pointer to AuthBackendRegistry is itself an AuthBackend that can be
// used in another AuthBackendRegistry.
type AuthBackendRegistry struct {
	backends []AuthBackend
}

// RegisterBackend registers the given backend in this registry.
// The newly added backend is inserted at the top of the list, so
// that it will override any existing backend that already manages
// the same uids.
func (ar *AuthBackendRegistry) RegisterBackend(backend AuthBackend) {
	ar.backends = append([]AuthBackend{backend}, ar.backends...)
}

// Authenticate tries to authenticate the user with the given uid and secret.
// Backends are polled in order. The user is authenticated as soon as one
// backend authenticates his uid with the given secret.
func (ar *AuthBackendRegistry) Authenticate(login, secret string, context *types.Context) (int64, error) {
	for _, backend := range ar.backends {
		uid, err := backend.Authenticate(login, secret, context)
		if err != nil {
			switch err.(type) {
			case UserNotFoundError:
				continue
			case InvalidCredentialsError:
				return 0, err
			}
		}
		return uid, nil
	}
	return 0, UserNotFoundError(login)
}

var _ AuthBackend = new(AuthBackendRegistry)
