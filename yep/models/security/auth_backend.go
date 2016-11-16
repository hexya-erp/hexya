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

// AuthenticationRegistry is the authentication registry of the application
var AuthenticationRegistry *AuthBackendRegistry

// An AuthBackend is an interface that is capable of authenticating a
// user and tell whether a user is a member of a given group.
type AuthBackend interface {
	// Authenticate returns true if the given secret authenticates the
	// user with the given uid. It returns false if the secret is not
	// valid or if the given uid is not known to this auth backend.
	Authenticate(uid int64, secret string) bool
	// UserGroups returns the list of groups the given uid belongs to.
	// It returns an empty slice if the user is not part of any group.
	// It returns nil if the given uid is not known to this auth backend.
	UserGroups(uid int64) []*Group
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
func (ar *AuthBackendRegistry) Authenticate(uid int64, secret string) bool {
	for _, backend := range ar.backends {
		if backend.Authenticate(uid, secret) {
			return true
		}
	}
	return false
}

// UserGroups returns the list of groups the given uid belongs to.
// Backends are polled in order. The returned list of groups is that
// of the first backend which replied with a non nil answer.
func (ar *AuthBackendRegistry) UserGroups(uid int64) []*Group {
	for _, backend := range ar.backends {
		if ug := backend.UserGroups(uid); ug != nil {
			return ug
		}
	}
	return nil
}

var _ AuthBackend = new(AuthBackendRegistry)

// AdminOnlyBackend is a simple backend that:
// - Never authenticate any user
// - States that Admin user is member of the admin group
// - States that any other user is member of no groups
// It is the default backend of the application.
type AdminOnlyBackend bool

// Authenticate function of the AdminOnlyBackend. Always returns false.
func (aob AdminOnlyBackend) Authenticate(uid int64, secret string) bool {
	return false
}

// UserGroups function of the AdminOnlyBackend. Returns the Admin group.
func (aob AdminOnlyBackend) UserGroups(uid int64) []*Group {
	if uid == SuperUserID {
		return []*Group{AdminGroup}
	}
	return []*Group{}
}

var _ AuthBackend = AdminOnlyBackend(true)

// GroupMapBackend is a simple backend that holds group / user mappings.
// It is mainly used for tests and for security reasons never authenticate
// any user.
type GroupMapBackend map[int64][]*Group

// Authenticate function of the GroupMapBackend. Always returns false.
func (gmb GroupMapBackend) Authenticate(uid int64, secret string) bool {
	return false
}

// UserGroups function of the GroupMapBackend.
// Returns the group slice for the given uid.
func (gmb GroupMapBackend) UserGroups(uid int64) []*Group {
	return gmb[uid]
}

var _ AuthBackend = GroupMapBackend{}
