// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

//go:generate ./generate-mock.sh

package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
)

// APIInterface wraps the
type APIInterface interface {
	AdminBlockUser(ctx context.Context, req *AdminBlockUserRequest) (err error)
	GetLoggedInUser(ctx context.Context) (res *User, err error)
	UpdateLoggedInUser(ctx context.Context, user *User) (res *User, err error)
	GetAuthProviders(ctx context.Context) (res []*AuthProviderInfo, err error)
	GetOwnAuthProviders(ctx context.Context) (res []*AuthProviderEntry, err error)
	UpdateOwnAuthProvider(ctx context.Context, params *UpdateOwnAuthProviderParams) (err error)
	DeleteOwnAuthProvider(ctx context.Context, params *DeleteOwnAuthProviderParams) (err error)
	GetBranding(ctx context.Context) (res *Branding, err error)
	GetConfiguration(ctx context.Context) (res *Configuration, err error)
	GetBhojpurTokenScopes(ctx context.Context, tokenHash string) (res []string, err error)
	GetToken(ctx context.Context, query *GetTokenSearchOptions) (res *Token, err error)
	GetPortAuthenticationToken(ctx context.Context, applicationID string) (res *Token, err error)
	DeleteAccount(ctx context.Context) (err error)
	GetClientRegion(ctx context.Context) (res string, err error)
	HasPermission(ctx context.Context, permission *PermissionName) (res bool, err error)
	GetApplications(ctx context.Context, options *GetApplicationsOptions) (res []*ApplicationInfo, err error)
	GetApplicationOwner(ctx context.Context, applicationID string) (res *UserInfo, err error)
	GetApplicationUsers(ctx context.Context, applicationID string) (res []*ApplicationInstanceUser, err error)
	GetFeaturedRepositories(ctx context.Context) (res []*WhitelistedRepository, err error)
	GetApplication(ctx context.Context, id string) (res *ApplicationInfo, err error)
	IsApplicationOwner(ctx context.Context, applicationID string) (res bool, err error)
	CreateApplication(ctx context.Context, options *CreateApplicationOptions) (res *ApplicationCreationResult, err error)
	StartApplication(ctx context.Context, id string, options *StartApplicationOptions) (res *StartApplicationResult, err error)
	StopApplication(ctx context.Context, id string) (err error)
	DeleteApplication(ctx context.Context, id string) (err error)
	SetApplicationDescription(ctx context.Context, id string, desc string) (err error)
	ControlAdmission(ctx context.Context, id string, level *AdmissionLevel) (err error)
	UpdateApplicationUserPin(ctx context.Context, id string, action *PinAction) (err error)
	SendHeartBeat(ctx context.Context, options *SendHeartBeatOptions) (err error)
	WatchApplicationImageBuildLogs(ctx context.Context, applicationID string) (err error)
	IsPrebuildDone(ctx context.Context, pwsid string) (res bool, err error)
	SetApplicationTimeout(ctx context.Context, applicationID string, duration *ApplicationTimeoutDuration) (res *SetApplicationTimeoutResult, err error)
	GetApplicationTimeout(ctx context.Context, applicationID string) (res *GetApplicationTimeoutResult, err error)
	GetOpenPorts(ctx context.Context, applicationID string) (res []*ApplicationInstancePort, err error)
	OpenPort(ctx context.Context, applicationID string, port *ApplicationInstancePort) (res *ApplicationInstancePort, err error)
	ClosePort(ctx context.Context, applicationID string, port float32) (err error)
	GetUserStorageResource(ctx context.Context, options *GetUserStorageResourceOptions) (res string, err error)
	UpdateUserStorageResource(ctx context.Context, options *UpdateUserStorageResourceOptions) (err error)
	GetEnvVars(ctx context.Context) (res []*UserEnvVarValue, err error)
	SetEnvVar(ctx context.Context, variable *UserEnvVarValue) (err error)
	DeleteEnvVar(ctx context.Context, variable *UserEnvVarValue) (err error)
	GetContentBlobUploadURL(ctx context.Context, name string) (url string, err error)
	GetContentBlobDownloadURL(ctx context.Context, name string) (url string, err error)
	GetBhojpurTokens(ctx context.Context) (res []*APIToken, err error)
	GenerateNewBhojpurToken(ctx context.Context, options *GenerateNewBhojpurTokenOptions) (res string, err error)
	DeleteBhojpurToken(ctx context.Context, tokenHash string) (err error)
	SendFeedback(ctx context.Context, feedback string) (res string, err error)
	RegisterGithubApp(ctx context.Context, installationID string) (err error)
	TakeSnapshot(ctx context.Context, options *TakeSnapshotOptions) (res string, err error)
	WaitForSnapshot(ctx context.Context, snapshotId string) (err error)
	GetSnapshots(ctx context.Context, applicationID string) (res []*string, err error)
	StoreLayout(ctx context.Context, applicationID string, layoutData string) (err error)
	GetLayout(ctx context.Context, applicationID string) (res string, err error)
	PreparePluginUpload(ctx context.Context, params *PreparePluginUploadParams) (res string, err error)
	ResolvePlugins(ctx context.Context, applicationID string, params *ResolvePluginsParams) (res *ResolvedPlugins, err error)
	InstallUserPlugins(ctx context.Context, params *InstallPluginsParams) (res bool, err error)
	UninstallUserPlugin(ctx context.Context, params *UninstallPluginParams) (res bool, err error)
	GuessGitTokenScopes(ctx context.Context, params *GuessGitTokenScopesParams) (res *GuessedGitTokenScopes, err error)

	InstanceUpdates(ctx context.Context, instanceID string) (<-chan *ApplicationInstance, error)
}

// FunctionName is the name of an RPC function
type FunctionName string

const (
	// FunctionAdminBlockUser is the name of the adminBlockUser function
	FunctionAdminBlockUser FunctionName = "adminBlockUser"
	// FunctionGetLoggedInUser is the name of the getLoggedInUser function
	FunctionGetLoggedInUser FunctionName = "getLoggedInUser"
	// FunctionUpdateLoggedInUser is the name of the updateLoggedInUser function
	FunctionUpdateLoggedInUser FunctionName = "updateLoggedInUser"
	// FunctionGetAuthProviders is the name of the getAuthProviders function
	FunctionGetAuthProviders FunctionName = "getAuthProviders"
	// FunctionGetOwnAuthProviders is the name of the getOwnAuthProviders function
	FunctionGetOwnAuthProviders FunctionName = "getOwnAuthProviders"
	// FunctionUpdateOwnAuthProvider is the name of the updateOwnAuthProvider function
	FunctionUpdateOwnAuthProvider FunctionName = "updateOwnAuthProvider"
	// FunctionDeleteOwnAuthProvider is the name of the deleteOwnAuthProvider function
	FunctionDeleteOwnAuthProvider FunctionName = "deleteOwnAuthProvider"
	// FunctionGetBranding is the name of the getBranding function
	FunctionGetBranding FunctionName = "getBranding"
	// FunctionGetConfiguration is the name of the getConfiguration function
	FunctionGetConfiguration FunctionName = "getConfiguration"
	// FunctionGetBhojpurTokenScopes is the name of the GetBhojpurTokenScopes function
	FunctionGetBhojpurTokenScopes FunctionName = "getBhojpurTokenScopes"
	// FunctionGetToken is the name of the getToken function
	FunctionGetToken FunctionName = "getToken"
	// FunctionGetPortAuthenticationToken is the name of the getPortAuthenticationToken function
	FunctionGetPortAuthenticationToken FunctionName = "getPortAuthenticationToken"
	// FunctionDeleteAccount is the name of the deleteAccount function
	FunctionDeleteAccount FunctionName = "deleteAccount"
	// FunctionGetClientRegion is the name of the getClientRegion function
	FunctionGetClientRegion FunctionName = "getClientRegion"
	// FunctionHasPermission is the name of the hasPermission function
	FunctionHasPermission FunctionName = "hasPermission"
	// FunctionGetApplications is the name of the getApplications function
	FunctionGetApplications FunctionName = "getApplications"
	// FunctionGetApplicationOwner is the name of the getApplicationOwner function
	FunctionGetApplicationOwner FunctionName = "getApplicationOwner"
	// FunctionGetApplicationUsers is the name of the getApplicationUsers function
	FunctionGetApplicationUsers FunctionName = "getApplicationUsers"
	// FunctionGetFeaturedRepositories is the name of the getFeaturedRepositories function
	FunctionGetFeaturedRepositories FunctionName = "getFeaturedRepositories"
	// FunctionGetApplication is the name of the getApplication function
	FunctionGetApplication FunctionName = "getApplication"
	// FunctionIsApplicationOwner is the name of the isApplicationOwner function
	FunctionIsApplicationOwner FunctionName = "isApplicationOwner"
	// FunctionCreateApplication is the name of the createApplication function
	FunctionCreateApplication FunctionName = "createApplication"
	// FunctionStartApplication is the name of the startApplication function
	FunctionStartApplication FunctionName = "startApplication"
	// FunctionStopApplication is the name of the stopApplication function
	FunctionStopApplication FunctionName = "stopApplication"
	// FunctionDeleteApplication is the name of the deleteApplication function
	FunctionDeleteApplication FunctionName = "deleteApplication"
	// FunctionSetApplicationDescription is the name of the setApplicationDescription function
	FunctionSetApplicationDescription FunctionName = "setApplicationDescription"
	// FunctionControlAdmission is the name of the controlAdmission function
	FunctionControlAdmission FunctionName = "controlAdmission"
	// FunctionUpdateApplicationUserPin is the name of the updateApplicationUserPin function
	FunctionUpdateApplicationUserPin FunctionName = "updateApplicationUserPin"
	// FunctionSendHeartBeat is the name of the sendHeartBeat function
	FunctionSendHeartBeat FunctionName = "sendHeartBeat"
	// FunctionWatchApplicationImageBuildLogs is the name of the watchApplicationImageBuildLogs function
	FunctionWatchApplicationImageBuildLogs FunctionName = "watchApplicationImageBuildLogs"
	// FunctionIsPrebuildDone is the name of the isPrebuildDone function
	FunctionIsPrebuildDone FunctionName = "isPrebuildDone"
	// FunctionSetApplicationTimeout is the name of the setApplicationTimeout function
	FunctionSetApplicationTimeout FunctionName = "setApplicationTimeout"
	// FunctionGetApplicationTimeout is the name of the getApplicationTimeout function
	FunctionGetApplicationTimeout FunctionName = "getApplicationTimeout"
	// FunctionGetOpenPorts is the name of the getOpenPorts function
	FunctionGetOpenPorts FunctionName = "getOpenPorts"
	// FunctionOpenPort is the name of the openPort function
	FunctionOpenPort FunctionName = "openPort"
	// FunctionClosePort is the name of the closePort function
	FunctionClosePort FunctionName = "closePort"
	// FunctionGetUserStorageResource is the name of the getUserStorageResource function
	FunctionGetUserStorageResource FunctionName = "getUserStorageResource"
	// FunctionUpdateUserStorageResource is the name of the updateUserStorageResource function
	FunctionUpdateUserStorageResource FunctionName = "updateUserStorageResource"
	// FunctionGetEnvVars is the name of the getEnvVars function
	FunctionGetEnvVars FunctionName = "getEnvVars"
	// FunctionSetEnvVar is the name of the setEnvVar function
	FunctionSetEnvVar FunctionName = "setEnvVar"
	// FunctionDeleteEnvVar is the name of the deleteEnvVar function
	FunctionDeleteEnvVar FunctionName = "deleteEnvVar"
	// FunctionGetContentBlobUploadURL is the name fo the getContentBlobUploadUrl function
	FunctionGetContentBlobUploadURL FunctionName = "getContentBlobUploadUrl"
	// FunctionGetContentBlobDownloadURL is the name fo the getContentBlobDownloadUrl function
	FunctionGetContentBlobDownloadURL FunctionName = "getContentBlobDownloadUrl"
	// FunctionGetBhojpurTokens is the name of the getBhojpurTokens function
	FunctionGetBhojpurTokens FunctionName = "getBhojpurTokens"
	// FunctionGenerateNewBhojpurToken is the name of the generateNewBhojpurToken function
	FunctionGenerateNewBhojpurToken FunctionName = "generateNewBhojpurToken"
	// FunctionDeleteBhojpurToken is the name of the deleteBhojpurToken function
	FunctionDeleteBhojpurToken FunctionName = "deleteBhojpurToken"
	// FunctionSendFeedback is the name of the sendFeedback function
	FunctionSendFeedback FunctionName = "sendFeedback"
	// FunctionRegisterGithubApp is the name of the registerGithubApp function
	FunctionRegisterGithubApp FunctionName = "registerGithubApp"
	// FunctionTakeSnapshot is the name of the takeSnapshot function
	FunctionTakeSnapshot FunctionName = "takeSnapshot"
	// FunctionGetSnapshots is the name of the getSnapshots function
	FunctionGetSnapshots FunctionName = "getSnapshots"
	// FunctionStoreLayout is the name of the storeLayout function
	FunctionStoreLayout FunctionName = "storeLayout"
	// FunctionGetLayout is the name of the getLayout function
	FunctionGetLayout FunctionName = "getLayout"
	// FunctionPreparePluginUpload is the name of the preparePluginUpload function
	FunctionPreparePluginUpload FunctionName = "preparePluginUpload"
	// FunctionResolvePlugins is the name of the resolvePlugins function
	FunctionResolvePlugins FunctionName = "resolvePlugins"
	// FunctionInstallUserPlugins is the name of the installUserPlugins function
	FunctionInstallUserPlugins FunctionName = "installUserPlugins"
	// FunctionUninstallUserPlugin is the name of the uninstallUserPlugin function
	FunctionUninstallUserPlugin FunctionName = "uninstallUserPlugin"
	// FunctionGuessGitTokenScopes is the name of the guessGitTokenScopes function
	FunctionGuessGitTokenScope FunctionName = "guessGitTokenScopes"

	// FunctionOnInstanceUpdate is the name of the onInstanceUpdate callback function
	FunctionOnInstanceUpdate = "onInstanceUpdate"
)

var errNotConnected = errors.New("not connected to Bhojpur server")

// ConnectToServerOpts configures the server connection
type ConnectToServerOpts struct {
	Context             context.Context
	Token               string
	Log                 *logrus.Entry
	ReconnectionHandler func()
	CloseHandler        func(error)
	ExtraHeaders        map[string]string
}

// ConnectToServer establishes a new websocket connection to the server
func ConnectToServer(endpoint string, opts ConnectToServerOpts) (*APIoverJSONRPC, error) {
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	epURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, xerrors.Errorf("invalid endpoint URL: %w", err)
	}

	var protocol string
	if epURL.Scheme == "wss:" {
		protocol = "https"
	} else {
		protocol = "http"
	}
	origin := fmt.Sprintf("%s://%s/", protocol, epURL.Hostname())

	reqHeader := http.Header{}
	reqHeader.Set("Origin", origin)
	for k, v := range opts.ExtraHeaders {
		reqHeader.Set(k, v)
	}
	if opts.Token != "" {
		reqHeader.Set("Authorization", "Bearer "+opts.Token)
	}
	ws := NewReconnectingWebsocket(endpoint, reqHeader, opts.Log)
	ws.ReconnectionHandler = opts.ReconnectionHandler
	go func() {
		err := ws.Dial(opts.Context)
		if opts.CloseHandler != nil {
			opts.CloseHandler(err)
		}
	}()

	var res APIoverJSONRPC
	res.log = opts.Log
	res.C = jsonrpc2.NewConn(opts.Context, ws, jsonrpc2.HandlerWithError(res.handler))
	return &res, nil
}

// APIoverJSONRPC makes JSON RPC calls to the Bhojpur.NET Platform server is the APIoverJSONRPC message type
type APIoverJSONRPC struct {
	C   jsonrpc2.JSONRPC2
	log *logrus.Entry

	mu   sync.RWMutex
	subs map[string]map[chan *ApplicationInstance]struct{}
}

// Close closes the connection
func (bp *APIoverJSONRPC) Close() (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	e1 := bp.C.Close()
	if e1 != nil {
		return e1
	}
	return nil
}

// InstanceUpdates subscribes to application instance updates until the context is canceled or the application
// instance is stopped.
func (bp *APIoverJSONRPC) InstanceUpdates(ctx context.Context, instanceID string) (<-chan *ApplicationInstance, error) {
	if bp == nil {
		return nil, errNotConnected
	}
	chn := make(chan *ApplicationInstance)

	bp.mu.Lock()
	if bp.subs == nil {
		bp.subs = make(map[string]map[chan *ApplicationInstance]struct{})
	}
	if sub, ok := bp.subs[instanceID]; ok {
		sub[chn] = struct{}{}
	} else {
		bp.subs[instanceID] = map[chan *ApplicationInstance]struct{}{chn: {}}
	}
	bp.mu.Unlock()

	go func() {
		<-ctx.Done()

		bp.mu.Lock()
		delete(bp.subs[instanceID], chn)
		close(chn)
		bp.mu.Unlock()
	}()

	return chn, nil
}

func (bp *APIoverJSONRPC) handler(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	if req.Method != FunctionOnInstanceUpdate {
		return
	}

	var instance ApplicationInstance
	err = json.Unmarshal(*req.Params, &instance)
	if err != nil {
		bp.log.WithError(err).WithField("raw", string(*req.Params)).Error("cannot unmarshal instance update")
		return
	}

	bp.mu.RLock()
	defer bp.mu.RUnlock()
	for chn := range bp.subs[instance.ID] {
		select {
		case chn <- &instance:
		default:
		}
	}
	for chn := range bp.subs[""] {
		select {
		case chn <- &instance:
		default:
		}
	}

	return
}

// AdminBlockUser calls adminBlockUser on the server
func (bp *APIoverJSONRPC) AdminBlockUser(ctx context.Context, message *AdminBlockUserRequest) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}
	_params = append(_params, message)

	var _result interface{}
	err = bp.C.Call(ctx, "adminBlockUser", _params, &_result)
	if err != nil {
		return err
	}
	return
}

// GetLoggedInUser calls getLoggedInUser on the server
func (bp *APIoverJSONRPC) GetLoggedInUser(ctx context.Context) (res *User, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result User
	err = bp.C.Call(ctx, "getLoggedInUser", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// UpdateLoggedInUser calls updateLoggedInUser on the server
func (bp *APIoverJSONRPC) UpdateLoggedInUser(ctx context.Context, user *User) (res *User, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, user)

	var result User
	err = bp.C.Call(ctx, "updateLoggedInUser", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetAuthProviders calls getAuthProviders on the server
func (bp *APIoverJSONRPC) GetAuthProviders(ctx context.Context) (res []*AuthProviderInfo, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result []*AuthProviderInfo
	err = bp.C.Call(ctx, "getAuthProviders", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetOwnAuthProviders calls getOwnAuthProviders on the server
func (bp *APIoverJSONRPC) GetOwnAuthProviders(ctx context.Context) (res []*AuthProviderEntry, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result []*AuthProviderEntry
	err = bp.C.Call(ctx, "getOwnAuthProviders", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// UpdateOwnAuthProvider calls updateOwnAuthProvider on the server
func (bp *APIoverJSONRPC) UpdateOwnAuthProvider(ctx context.Context, params *UpdateOwnAuthProviderParams) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	err = bp.C.Call(ctx, "updateOwnAuthProvider", _params, nil)
	if err != nil {
		return
	}

	return
}

// DeleteOwnAuthProvider calls deleteOwnAuthProvider on the server
func (bp *APIoverJSONRPC) DeleteOwnAuthProvider(ctx context.Context, params *DeleteOwnAuthProviderParams) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	err = bp.C.Call(ctx, "deleteOwnAuthProvider", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetBranding calls getBranding on the server
func (bp *APIoverJSONRPC) GetBranding(ctx context.Context) (res *Branding, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result Branding
	err = bp.C.Call(ctx, "getBranding", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetConfiguration calls getConfiguration on the server
func (bp *APIoverJSONRPC) GetConfiguration(ctx context.Context) (res *Configuration, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result Configuration
	err = bp.C.Call(ctx, "getConfiguration", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetBhojpurTokenScopes calls getBhojpurTokenScopes on the server
func (bp *APIoverJSONRPC) GetBhojpurTokenScopes(ctx context.Context, tokenHash string) (res []string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, tokenHash)

	var result []string
	err = bp.C.Call(ctx, "getBhojpurTokenScopes", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetToken calls getToken on the server
func (bp *APIoverJSONRPC) GetToken(ctx context.Context, query *GetTokenSearchOptions) (res *Token, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, query)

	var result Token
	err = bp.C.Call(ctx, "getToken", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetPortAuthenticationToken calls getPortAuthenticationToken on the server
func (bp *APIoverJSONRPC) GetPortAuthenticationToken(ctx context.Context, applicationID string) (res *Token, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result Token
	err = bp.C.Call(ctx, "getPortAuthenticationToken", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// DeleteAccount calls deleteAccount on the server
func (bp *APIoverJSONRPC) DeleteAccount(ctx context.Context) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	err = bp.C.Call(ctx, "deleteAccount", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetClientRegion calls getClientRegion on the server
func (bp *APIoverJSONRPC) GetClientRegion(ctx context.Context) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result string
	err = bp.C.Call(ctx, "getClientRegion", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// HasPermission calls hasPermission on the server
func (bp *APIoverJSONRPC) HasPermission(ctx context.Context, permission *PermissionName) (res bool, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, permission)

	var result bool
	err = bp.C.Call(ctx, "hasPermission", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetApplications calls getApplications on the server
func (bp *APIoverJSONRPC) GetApplications(ctx context.Context, options *GetApplicationsOptions) (res []*ApplicationInfo, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	var result []*ApplicationInfo
	err = bp.C.Call(ctx, "getApplications", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetApplicationOwner calls getApplicationOwner on the server
func (bp *APIoverJSONRPC) GetApplicationOwner(ctx context.Context, applicationID string) (res *UserInfo, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result UserInfo
	err = bp.C.Call(ctx, "getApplicationOwner", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetApplicationUsers calls getApplicationUsers on the server
func (bp *APIoverJSONRPC) GetApplicationUsers(ctx context.Context, applicationID string) (res []*ApplicationInstanceUser, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result []*ApplicationInstanceUser
	err = bp.C.Call(ctx, "getApplicationUsers", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetFeaturedRepositories calls getFeaturedRepositories on the server
func (bp *APIoverJSONRPC) GetFeaturedRepositories(ctx context.Context) (res []*WhitelistedRepository, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result []*WhitelistedRepository
	err = bp.C.Call(ctx, "getFeaturedRepositories", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GetApplication calls getApplication on the server
func (bp *APIoverJSONRPC) GetApplication(ctx context.Context, id string) (res *ApplicationInfo, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)

	var result ApplicationInfo
	err = bp.C.Call(ctx, "getApplication", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// IsApplicationOwner calls isApplicationOwner on the server
func (bp *APIoverJSONRPC) IsApplicationOwner(ctx context.Context, applicationID string) (res bool, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result bool
	err = bp.C.Call(ctx, "isApplicationOwner", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// CreateApplication calls createApplication on the server
func (bp *APIoverJSONRPC) CreateApplication(ctx context.Context, options *CreateApplicationOptions) (res *ApplicationCreationResult, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	var result ApplicationCreationResult
	err = bp.C.Call(ctx, "createApplication", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// StartApplication calls startApplication on the server
func (bp *APIoverJSONRPC) StartApplication(ctx context.Context, id string, options *StartApplicationOptions) (res *StartApplicationResult, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)
	_params = append(_params, options)

	var result StartApplicationResult
	err = bp.C.Call(ctx, "startApplication", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// StopApplication calls stopApplication on the server
func (bp *APIoverJSONRPC) StopApplication(ctx context.Context, id string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)

	err = bp.C.Call(ctx, "stopApplication", _params, nil)
	if err != nil {
		return
	}

	return
}

// DeleteApplication calls deleteApplication on the server
func (bp *APIoverJSONRPC) DeleteApplication(ctx context.Context, id string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)

	err = bp.C.Call(ctx, "deleteApplication", _params, nil)
	if err != nil {
		return
	}

	return
}

// SetApplicationDescription calls setApplicationDescription on the server
func (bp *APIoverJSONRPC) SetApplicationDescription(ctx context.Context, id string, desc string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)
	_params = append(_params, desc)

	err = bp.C.Call(ctx, "setApplicationDescription", _params, nil)
	if err != nil {
		return
	}

	return
}

// ControlAdmission calls controlAdmission on the server
func (bp *APIoverJSONRPC) ControlAdmission(ctx context.Context, id string, level *AdmissionLevel) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)
	_params = append(_params, level)

	err = bp.C.Call(ctx, "controlAdmission", _params, nil)
	if err != nil {
		return
	}

	return
}

// WatchApplicationImageBuildLogs calls watchApplicationImageBuildLogs on the server
func (bp *APIoverJSONRPC) WatchApplicationImageBuildLogs(ctx context.Context, applicationID string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	err = bp.C.Call(ctx, "watchApplicationImageBuildLogs", _params, nil)
	if err != nil {
		return
	}

	return
}

// IsPrebuildDone calls isPrebuildDone on the server
func (bp *APIoverJSONRPC) IsPrebuildDone(ctx context.Context, pwsid string) (res bool, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, pwsid)

	var result bool
	err = bp.C.Call(ctx, "isPrebuildDone", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// SetApplicationTimeout calls setApplicationTimeout on the server
func (bp *APIoverJSONRPC) SetApplicationTimeout(ctx context.Context, applicationID string, duration *ApplicationTimeoutDuration) (res *SetApplicationTimeoutResult, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)
	_params = append(_params, duration)

	var result SetApplicationTimeoutResult
	err = bp.C.Call(ctx, "setApplicationTimeout", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// GetApplicationTimeout calls getApplicationTimeout on the server
func (bp *APIoverJSONRPC) GetApplicationTimeout(ctx context.Context, ApplicationID string) (res *GetApplicationTimeoutResult, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result GetApplicationTimeoutResult
	err = bp.C.Call(ctx, "getApplicationTimeout", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// SendHeartBeat calls sendHeartBeat on the server
func (bp *APIoverJSONRPC) SendHeartBeat(ctx context.Context, options *SendHeartBeatOptions) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	err = bp.C.Call(ctx, "sendHeartBeat", _params, nil)
	if err != nil {
		return
	}

	return
}

// UpdateApplicationUserPin calls updateApplicationUserPin on the server
func (bp *APIoverJSONRPC) UpdateApplicationUserPin(ctx context.Context, id string, action *PinAction) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, id)
	_params = append(_params, action)

	err = bp.C.Call(ctx, "updateApplicationUserPin", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetOpenPorts calls getOpenPorts on the server
func (bp *APIoverJSONRPC) GetOpenPorts(ctx context.Context, applicationID string) (res []*ApplicationInstancePort, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result []*ApplicationInstancePort
	err = bp.C.Call(ctx, "getOpenPorts", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// OpenPort calls openPort on the server
func (bp *APIoverJSONRPC) OpenPort(ctx context.Context, applicationID string, port *ApplicationInstancePort) (res *ApplicationInstancePort, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)
	_params = append(_params, port)

	var result ApplicationInstancePort
	err = bp.C.Call(ctx, "openPort", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// ClosePort calls closePort on the server
func (bp *APIoverJSONRPC) ClosePort(ctx context.Context, applicationID string, port float32) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)
	_params = append(_params, port)

	err = bp.C.Call(ctx, "closePort", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetUserStorageResource calls getUserStorageResource on the server
func (bp *APIoverJSONRPC) GetUserStorageResource(ctx context.Context, options *GetUserStorageResourceOptions) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	var result string
	err = bp.C.Call(ctx, "getUserStorageResource", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// UpdateUserStorageResource calls updateUserStorageResource on the server
func (bp *APIoverJSONRPC) UpdateUserStorageResource(ctx context.Context, options *UpdateUserStorageResourceOptions) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	err = bp.C.Call(ctx, "updateUserStorageResource", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetEnvVars calls getEnvVars on the server
func (bp *APIoverJSONRPC) GetEnvVars(ctx context.Context) (res []*UserEnvVarValue, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result []*UserEnvVarValue
	err = bp.C.Call(ctx, "getEnvVars", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// SetEnvVar calls setEnvVar on the server
func (bp *APIoverJSONRPC) SetEnvVar(ctx context.Context, variable *UserEnvVarValue) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, variable)

	err = bp.C.Call(ctx, "setEnvVar", _params, nil)
	if err != nil {
		return
	}

	return
}

// DeleteEnvVar calls deleteEnvVar on the server
func (bp *APIoverJSONRPC) DeleteEnvVar(ctx context.Context, variable *UserEnvVarValue) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, variable)

	err = bp.C.Call(ctx, "deleteEnvVar", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetContentBlobUploadURL calls getContentBlobUploadUrl on the server
func (bp *APIoverJSONRPC) GetContentBlobUploadURL(ctx context.Context, name string) (url string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, name)

	var result string
	err = bp.C.Call(ctx, string(FunctionGetContentBlobUploadURL), _params, &result)
	if err != nil {
		return
	}
	url = result

	return
}

// GetContentBlobDownloadURL calls getContentBlobDownloadUrl on the server
func (bp *APIoverJSONRPC) GetContentBlobDownloadURL(ctx context.Context, name string) (url string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, name)

	var result string
	err = bp.C.Call(ctx, string(FunctionGetContentBlobDownloadURL), _params, &result)
	if err != nil {
		return
	}
	url = result

	return
}

// GetBhojpurTokens calls getBhojpurTokens on the server
func (bp *APIoverJSONRPC) GetBhojpurTokens(ctx context.Context) (res []*APIToken, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	var result []*APIToken
	err = bp.C.Call(ctx, "getBhojpurTokens", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GenerateNewBhojpurToken calls generateNewBhojpurToken on the server
func (bp *APIoverJSONRPC) GenerateNewBhojpurToken(ctx context.Context, options *GenerateNewBhojpurTokenOptions) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	var result string
	err = bp.C.Call(ctx, "generateNewBhojpurToken", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// DeleteBhojpurToken calls deleteBhojpurToken on the server
func (bp *APIoverJSONRPC) DeleteBhojpurToken(ctx context.Context, tokenHash string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, tokenHash)

	err = bp.C.Call(ctx, "deleteBhojpurToken", _params, nil)
	if err != nil {
		return
	}

	return
}

// SendFeedback calls sendFeedback on the server
func (bp *APIoverJSONRPC) SendFeedback(ctx context.Context, feedback string) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, feedback)

	var result string
	err = bp.C.Call(ctx, "sendFeedback", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// RegisterGithubApp calls registerGithubApp on the server
func (bp *APIoverJSONRPC) RegisterGithubApp(ctx context.Context, installationID string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, installationID)

	err = bp.C.Call(ctx, "registerGithubApp", _params, nil)
	if err != nil {
		return
	}

	return
}

// TakeSnapshot calls takeSnapshot on the server
func (bp *APIoverJSONRPC) TakeSnapshot(ctx context.Context, options *TakeSnapshotOptions) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, options)

	var result string
	err = bp.C.Call(ctx, "takeSnapshot", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// WaitForSnapshot calls waitForSnapshot on the server
func (bp *APIoverJSONRPC) WaitForSnapshot(ctx context.Context, snapshotId string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, snapshotId)

	var result string
	err = bp.C.Call(ctx, "waitForSnapshot", _params, &result)
	return
}

// GetSnapshots calls getSnapshots on the server
func (bp *APIoverJSONRPC) GetSnapshots(ctx context.Context, applicationID string) (res []*string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result []*string
	err = bp.C.Call(ctx, "getSnapshots", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// StoreLayout calls storeLayout on the server
func (bp *APIoverJSONRPC) StoreLayout(ctx context.Context, applicationID string, layoutData string) (err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)
	_params = append(_params, layoutData)

	err = bp.C.Call(ctx, "storeLayout", _params, nil)
	if err != nil {
		return
	}

	return
}

// GetLayout calls getLayout on the server
func (bp *APIoverJSONRPC) GetLayout(ctx context.Context, applicationID string) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)

	var result string
	err = bp.C.Call(ctx, "getLayout", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// PreparePluginUpload calls preparePluginUpload on the server
func (bp *APIoverJSONRPC) PreparePluginUpload(ctx context.Context, params *PreparePluginUploadParams) (res string, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	var result string
	err = bp.C.Call(ctx, "preparePluginUpload", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// ResolvePlugins calls resolvePlugins on the server
func (bp *APIoverJSONRPC) ResolvePlugins(ctx context.Context, applicationID string, params *ResolvePluginsParams) (res *ResolvedPlugins, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, applicationID)
	_params = append(_params, params)

	var result ResolvedPlugins
	err = bp.C.Call(ctx, "resolvePlugins", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// InstallUserPlugins calls installUserPlugins on the server
func (bp *APIoverJSONRPC) InstallUserPlugins(ctx context.Context, params *InstallPluginsParams) (res bool, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	var result bool
	err = bp.C.Call(ctx, "installUserPlugins", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// UninstallUserPlugin calls uninstallUserPlugin on the server
func (bp *APIoverJSONRPC) UninstallUserPlugin(ctx context.Context, params *UninstallPluginParams) (res bool, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	var result bool
	err = bp.C.Call(ctx, "uninstallUserPlugin", _params, &result)
	if err != nil {
		return
	}
	res = result

	return
}

// GuessGitTokenScopes calls GuessGitTokenScopes on the server
func (bp *APIoverJSONRPC) GuessGitTokenScopes(ctx context.Context, params *GuessGitTokenScopesParams) (res *GuessedGitTokenScopes, err error) {
	if bp == nil {
		err = errNotConnected
		return
	}
	var _params []interface{}

	_params = append(_params, params)

	var result GuessedGitTokenScopes
	err = bp.C.Call(ctx, "guessGitTokenScopes", _params, &result)
	if err != nil {
		return
	}
	res = &result

	return
}

// PermissionName is the name of a permission
type PermissionName string

const (
	// PermissionNameMonitor is the "monitor" permission
	PermissionNameMonitor PermissionName = "monitor"
	// PermissionNameEnforcement is the "enforcement" permission
	PermissionNameEnforcement PermissionName = "enforcement"
	// PermissionNamePrivilegedWs is the "privileged-ws" permission
	PermissionNamePrivilegedWs PermissionName = "privileged-ws"
	// PermissionNameRegistryAccess is the "registry-access" permission
	PermissionNameRegistryAccess PermissionName = "registry-access"
	// PermissionNameAdminUsers is the "admin-users" permission
	PermissionNameAdminUsers PermissionName = "admin-users"
	// PermissionNameAdminApplications is the "admin-applicaions" permission
	PermissionNameAdminApplications PermissionName = "admin-applications"
	// PermissionNameAdminAPI is the "admin-api" permission
	PermissionNameAdminAPI PermissionName = "admin-api"
)

// AdmissionLevel is the admission level to an application
type AdmissionLevel string

const (
	// AdmissionLevelOwner is the "owner" admission level
	AdmissionLevelOwner AdmissionLevel = "owner"
	// AdmissionLevelEveryone is the "everyone" admission level
	AdmissionLevelEveryone AdmissionLevel = "everyone"
)

// PinAction is the pin action
type PinAction string

const (
	// PinActionPin is the "pin" action
	PinActionPin PinAction = "pin"
	// PinActionUnpin is the "unpin" action
	PinActionUnpin PinAction = "unpin"
	// PinActionToggle is the "toggle" action
	PinActionToggle PinAction = "toggle"
)

// ApplicationTimeoutDuration is the durations one have set for the application timeout
type ApplicationTimeoutDuration string

const (
	// ApplicationTimeoutDuration30m sets "30m" as timeout duration
	ApplicationTimeoutDuration30m = "30m"
	// ApplicationTimeoutDuration60m sets "60m" as timeout duration
	ApplicationTimeoutDuration60m = "60m"
	// ApplicationTimeoutDuration180m sets "180m" as timeout duration
	ApplicationTimeoutDuration180m = "180m"
)

// UserInfo is the UserInfo message type
type UserInfo struct {
	Name string `json:"name,omitempty"`
}

// GetUserStorageResourceOptions is the GetUserStorageResourceOptions message type
type GetUserStorageResourceOptions struct {
	URI string `json:"uri,omitempty"`
}

// GetApplicationsOptions is the GetApplicationsOptions message type
type GetApplicationsOptions struct {
	Limit        float64 `json:"limit,omitempty"`
	PinnedOnly   bool    `json:"pinnedOnly,omitempty"`
	SearchString string  `json:"searchString,omitempty"`
}

// StartApplicationResult is the StartApplicationResult message type
type StartApplicationResult struct {
	InstanceID   string `json:"instanceID,omitempty"`
	ApplicationURL string `json:"applicationURL,omitempty"`
}

// APIToken is the APIToken message type
type APIToken struct {

	// Created timestamp
	Created string `json:"created,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`

	// Human readable name of the token
	Name string `json:"name,omitempty"`

	// Scopes (e.g. limition to read-only)
	Scopes []string `json:"scopes,omitempty"`

	// Hash value (SHA256) of the token (primary key).
	TokenHash string `json:"tokenHash,omitempty"`

	// // Token kindfloat64 is the float64 message type
	Type float64 `json:"type,omitempty"`

	// The user the token belongs to.
	User *User `json:"user,omitempty"`
}

// InstallPluginsParams is the InstallPluginsParams message type
type InstallPluginsParams struct {
	PluginIds []string `json:"pluginIds,omitempty"`
}

// OAuth2Config is the OAuth2Config message type
type OAuth2Config struct {
	AuthorizationParams map[string]string `json:"authorizationParams,omitempty"`
	AuthorizationURL    string            `json:"authorizationUrl,omitempty"`
	CallBackURL         string            `json:"callBackUrl,omitempty"`
	ClientID            string            `json:"clientId,omitempty"`
	ClientSecret        string            `json:"clientSecret,omitempty"`
	ConfigURL           string            `json:"configURL,omitempty"`
	Scope               string            `json:"scope,omitempty"`
	ScopeSeparator      string            `json:"scopeSeparator,omitempty"`
	SettingsURL         string            `json:"settingsUrl,omitempty"`
	TokenURL            string            `json:"tokenUrl,omitempty"`
}

// AuthProviderEntry is the AuthProviderEntry message type
type AuthProviderEntry struct {
	Host    string        `json:"host,omitempty"`
	ID      string        `json:"id,omitempty"`
	Oauth   *OAuth2Config `json:"oauth,omitempty"`
	OwnerID string        `json:"ownerId,omitempty"`

	// Status  string        `json:"status,omitempty"`   string is the    string message type
	Type string `json:"type,omitempty"`
}

// Commit is the Commit message type
type Commit struct {
	Ref        string      `json:"ref,omitempty"`
	RefType    string      `json:"refType,omitempty"`
	Repository *Repository `json:"repository,omitempty"`
	Revision   string      `json:"revision,omitempty"`
}

// Fork is the Fork message type
type Fork struct {
	Parent *Repository `json:"parent,omitempty"`
}

// Repository is the Repository message type
type Repository struct {
	AvatarURL     string `json:"avatarUrl,omitempty"`
	CloneURL      string `json:"cloneUrl,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	Description   string `json:"description,omitempty"`
	Fork          *Fork  `json:"fork,omitempty"`
	Host          string `json:"host,omitempty"`
	Name          string `json:"name,omitempty"`
	Owner         string `json:"owner,omitempty"`

	// Optional for backwards compatibility
	Private bool   `json:"private,omitempty"`
	WebURL  string `json:"webUrl,omitempty"`
}

// ApplicationCreationResult is the ApplicationCreationResult message type
type ApplicationCreationResult struct {
	CreatedApplicationID         string                    `json:"createdApplicationId,omitempty"`
	ExistingApplications         []*ApplicationInfo        `json:"existingApplications,omitempty"`
	RunningPrebuildApplicationID string                    `json:"runningPrebuildApplicationID,omitempty"`
	RunningApplicationPrebuild   *RunningApplicationPrebuild `json:"runningApplicationPrebuild,omitempty"`
	ApplicationURL               string                    `json:"applicationURL,omitempty"`
}

// RunningApplicationPrebuild is the RunningApplicationPrebuild message type
type RunningApplicationPrebuild struct {
	PrebuildID  string `json:"prebuildID,omitempty"`
	SameCluster bool   `json:"sameCluster,omitempty"`
	Starting    string `json:"starting,omitempty"`
	ApplicationID string `json:"applicationID,omitempty"`
}

// Application is the Bhojpur.NET Platform application message type
type Application struct {

	// The resolved/built fixed named of the base image. This field is only set if the application
	// already has its base image built.
	BaseImageNameResolved string           `json:"baseImageNameResolved,omitempty"`
	BasedOnPrebuildID     string           `json:"basedOnPrebuildId,omitempty"`
	BasedOnSnapshotID     string           `json:"basedOnSnapshotId,omitempty"`
	Config                *ApplicationConfig `json:"config,omitempty"`

	// Marks the time when the application content has been deleted.
	ContentDeletedTime string            `json:"contentDeletedTime,omitempty"`
	Context            *ApplicationContext `json:"context,omitempty"`
	ContextURL         string            `json:"contextURL,omitempty"`
	CreationTime       string            `json:"creationTime,omitempty"`
	Deleted            bool              `json:"deleted,omitempty"`
	Description        string            `json:"description,omitempty"`
	ID                 string            `json:"id,omitempty"`

	// The resolved, fix name of the application image. We only use this
	// to access the logs during an image build.
	ImageNameResolved string `json:"imageNameResolved,omitempty"`

	// The source where to get the application base image from. This source is resolved
	// during application creation. Once a base image has been built the information in here
	// is superseded by baseImageNameResolved.
	ImageSource interface{} `json:"imageSource,omitempty"`
	OwnerID     string      `json:"ownerId,omitempty"`
	Pinned      bool        `json:"pinned,omitempty"`
	Shareable   bool        `json:"shareable,omitempty"`

	// Mark as deleted (user-facing). The actual deletion of the application content is executed
	// with a (configurable) delay
	SoftDeleted string `json:"softDeleted,omitempty"`

	// Marks the time when the application was marked as softDeleted. The actual deletion of the
	// application content happens after a configurable period

	// SoftDeletedTime string `json:"softDeletedTime,omitempty"`           string is the            string message type
	Type string `json:"type,omitempty"`
}

// ApplicationConfig is the ApplicationConfig message type
type ApplicationConfig struct {
	CheckoutLocation string `json:"checkoutLocation,omitempty"`

	// Set of automatically inferred feature flags. That's not something the user can set, but
	// that is set by Bhojpur.NET Platform at application creation time.
	FeatureFlags []string          `json:"_featureFlags,omitempty"`
	GitConfig    map[string]string `json:"gitConfig,omitempty"`
	Github       *GithubAppConfig  `json:"github,omitempty"`
	Ide          string            `json:"ide,omitempty"`
	Image        interface{}       `json:"image,omitempty"`

	// Where the config object originates from.
	//
	// repo - from the repository
	// platform-validated - from github.com/bhojpur/platform-validated
	// derived - computed based on analyzing the repository
	// default - our static catch-all default config
	Origin            string        `json:"_origin,omitempty"`
	Ports             []*PortConfig `json:"ports,omitempty"`
	Privileged        bool          `json:"privileged,omitempty"`
	Tasks             []*TaskConfig `json:"tasks,omitempty"`
	Vscode            *VSCodeConfig `json:"vscode,omitempty"`
	ApplicationLocation string      `json:"applicationLocation,omitempty"`
}

// ApplicationContext is the ApplicationContext message type
type ApplicationContext struct {
	ForceCreateNewApplication bool `json:"forceCreateNewApplication,omitempty"`
	NormalizedContextURL    string `json:"normalizedContextURL,omitempty"`
	Title                   string `json:"title,omitempty"`
}

// ApplicationImageSourceDocker is the ApplicationImageSourceDocker message type
type ApplicationImageSourceDocker struct {
	DockerFileHash   string  `json:"dockerFileHash,omitempty"`
	DockerFilePath   string  `json:"dockerFilePath,omitempty"`
	DockerFileSource *Commit `json:"dockerFileSource,omitempty"`
}

// ApplicationImageSourceReference is the ApplicationImageSourceReference message type
type ApplicationImageSourceReference struct {

	// The resolved, fix base image reference
	BaseImageResolved string `json:"baseImageResolved,omitempty"`
}

// ApplicationInfo is the ApplicationInfo message type
type ApplicationInfo struct {
	LatestInstance *ApplicationInstance `json:"latestInstance,omitempty"`
	Application    *Application         `json:"application,omitempty"`
}

// ApplicationInstance is the ApplicationInstance message type
type ApplicationInstance struct {
	Configuration  *ApplicationInstanceConfiguration `json:"configuration,omitempty"`
	CreationTime   string                          `json:"creationTime,omitempty"`
	Deleted        bool                            `json:"deleted,omitempty"`
	DeployedTime   string                          `json:"deployedTime,omitempty"`
	ID             string                          `json:"id,omitempty"`
	IdeURL         string                          `json:"ideUrl,omitempty"`
	Region         string                          `json:"region,omitempty"`
	StartedTime    string                          `json:"startedTime,omitempty"`
	Status         *ApplicationInstanceStatus      `json:"status,omitempty"`
	StoppedTime    string                          `json:"stoppedTime,omitempty"`
	ApplicationID    string                        `json:"applicationId,omitempty"`
	ApplicationImage string                        `json:"applicationImage,omitempty"`
}

// ApplicationInstanceConditions is the ApplicationInstanceConditions message type
type ApplicationInstanceConditions struct {
	Deployed          bool   `json:"deployed,omitempty"`
	Failed            string `json:"failed,omitempty"`
	FirstUserActivity string `json:"firstUserActivity,omitempty"`
	NeededImageBuild  bool   `json:"neededImageBuild,omitempty"`
	PullingImages     bool   `json:"pullingImages,omitempty"`
	Timeout           string `json:"timeout,omitempty"`
}

// ApplicationInstanceConfiguration is the ApplicationInstanceConfiguration message type
type ApplicationInstanceConfiguration struct {
	FeatureFlags []string `json:"featureFlags,omitempty"`
	TheiaVersion string   `json:"theiaVersion,omitempty"`
}

// ApplicationInstanceRepoStatus is the ApplicationInstanceRepoStatus message type
type ApplicationInstanceRepoStatus struct {
	Branch               string   `json:"branch,omitempty"`
	LatestCommit         string   `json:"latestCommit,omitempty"`
	TotalUncommitedFiles float64  `json:"totalUncommitedFiles,omitempty"`
	TotalUnpushedCommits float64  `json:"totalUnpushedCommits,omitempty"`
	TotalUntrackedFiles  float64  `json:"totalUntrackedFiles,omitempty"`
	UncommitedFiles      []string `json:"uncommitedFiles,omitempty"`
	UnpushedCommits      []string `json:"unpushedCommits,omitempty"`
	UntrackedFiles       []string `json:"untrackedFiles,omitempty"`
}

// ApplicationInstanceStatus is the ApplicationInstanceStatus message type
type ApplicationInstanceStatus struct {
	Conditions   *ApplicationInstanceConditions `json:"conditions,omitempty"`
	ExposedPorts []*ApplicationInstancePort   `json:"exposedPorts,omitempty"`
	Message      string                       `json:"message,omitempty"`
	NodeName     string                       `json:"nodeName,omitempty"`
	OwnerToken   string                       `json:"ownerToken,omitempty"`
	Phase        string                       `json:"phase,omitempty"`
	Repo         *ApplicationInstanceRepoStatus `json:"repo,omitempty"`
	Timeout      string                       `json:"timeout,omitempty"`
}

// StartApplicationOptions is the StartApplicationOptions message type
type StartApplicationOptions struct {
	ForceDefaultImage bool `json:"forceDefaultImage,omitempty"`
}

// GetApplicationTimeoutResult is the GetApplicationTimeoutResult message type
type GetApplicationTimeoutResult struct {
	CanChange bool   `json:"canChange,omitempty"`
	Duration  string `json:"duration,omitempty"`
}

// ApplicationInstancePort is the ApplicationInstancePort message type
type ApplicationInstancePort struct {
	Port       float64 `json:"port,omitempty"`
	URL        string  `json:"url,omitempty"`
	Visibility string  `json:"visibility,omitempty"`
}

// GithubAppConfig is the GithubAppConfig message type
type GithubAppConfig struct {
	Prebuilds *GithubAppPrebuildConfig `json:"prebuilds,omitempty"`
}

// GithubAppPrebuildConfig is the GithubAppPrebuildConfig message type
type GithubAppPrebuildConfig struct {
	AddBadge              bool        `json:"addBadge,omitempty"`
	AddCheck              bool        `json:"addCheck,omitempty"`
	AddComment            bool        `json:"addComment,omitempty"`
	AddLabel              interface{} `json:"addLabel,omitempty"`
	Branches              bool        `json:"branches,omitempty"`
	Master                bool        `json:"master,omitempty"`
	PullRequests          bool        `json:"pullRequests,omitempty"`
	PullRequestsFromForks bool        `json:"pullRequestsFromForks,omitempty"`
}

// ImageConfigFile is the ImageConfigFile message type
type ImageConfigFile struct {
	Context string `json:"context,omitempty"`
	File    string `json:"file,omitempty"`
}

// PortConfig is the PortConfig message type
type PortConfig struct {
	OnOpen     string  `json:"onOpen,omitempty"`
	Port       float64 `json:"port,omitempty"`
	Visibility string  `json:"visibility,omitempty"`
}

// ResolvedPlugins is the ResolvedPlugins message type
type ResolvedPlugins struct {
	AdditionalProperties map[string]*ResolvedPlugin `json:"-,omitempty"`
}

// ResolvePluginsParams is the ResolvePluginsParams message type
type ResolvePluginsParams struct {
	Builtins *ResolvedPlugins `json:"builtins,omitempty"`
	Config   *ApplicationConfig `json:"config,omitempty"`
}

// TaskConfig is the TaskConfig message type
type TaskConfig struct {
	Before   string                 `json:"before,omitempty"`
	Command  string                 `json:"command,omitempty"`
	Env      map[string]interface{} `json:"env,omitempty"`
	Init     string                 `json:"init,omitempty"`
	Name     string                 `json:"name,omitempty"`
	OpenIn   string                 `json:"openIn,omitempty"`
	OpenMode string                 `json:"openMode,omitempty"`
	Prebuild string                 `json:"prebuild,omitempty"`
}

// VSCodeConfig is the VSCodeConfig message type
type VSCodeConfig struct {
	Extensions []string `json:"extensions,omitempty"`
}

// MarshalJSON marshals to JSON
func (strct *ResolvedPlugins) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	buf.WriteString("{")
	comma := false
	// Marshal any additional Properties
	for k, v := range strct.AdditionalProperties {
		if comma {
			buf.WriteString(",")
		}
		buf.WriteString(fmt.Sprintf("\"%s\":", k))
		tmp, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		buf.Write(tmp)
		comma = true
	}

	buf.WriteString("}")
	rv := buf.Bytes()
	return rv, nil
}

// UnmarshalJSON marshals from JSON
func (strct *ResolvedPlugins) UnmarshalJSON(b []byte) error {
	var jsonMap map[string]json.RawMessage
	if err := json.Unmarshal(b, &jsonMap); err != nil {
		return err
	}
	// parse all the defined properties
	for k, v := range jsonMap {
		switch k {
		default:
			// an additional "*ResolvedPlugin" value
			var additionalValue *ResolvedPlugin
			if err := json.Unmarshal([]byte(v), &additionalValue); err != nil {
				return err // invalid additionalProperty
			}
			if strct.AdditionalProperties == nil {
				strct.AdditionalProperties = make(map[string]*ResolvedPlugin)
			}
			strct.AdditionalProperties[k] = additionalValue
		}
	}
	return nil
}

// Configuration is the Configuration message type
type Configuration struct {
	DaysBeforeGarbageCollection float64 `json:"daysBeforeGarbageCollection,omitempty"`
	GarbageCollectionStartDate  float64 `json:"garbageCollectionStartDate,omitempty"`
}

// WhitelistedRepository is the WhitelistedRepository message type
type WhitelistedRepository struct {
	Avatar      string `json:"avatar,omitempty"`
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
}

// UserEnvVarValue is the UserEnvVarValue message type
type UserEnvVarValue struct {
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	RepositoryPattern string `json:"repositoryPattern,omitempty"`
	Value             string `json:"value,omitempty"`
}

// GenerateNewBhojpurTokenOptions is the GenerateNewBhojpurTokenOptions message type
type GenerateNewBhojpurTokenOptions struct {
	Name string `json:"name,omitempty"`

	// Scopes []string `json:"scopes,omitempty"`  float64 is the   float64 message type
	Type float64 `json:"type,omitempty"`
}

// TakeSnapshotOptions is the TakeSnapshotOptions message type
type TakeSnapshotOptions struct {
	LayoutData  string `json:"layoutData,omitempty"`
	ApplicationID string `json:"applicationId,omitempty"`
	DontWait    bool   `json:"dontWait",omitempty`
}

// PreparePluginUploadParams is the PreparePluginUploadParams message type
type PreparePluginUploadParams struct {
	FullPluginName string `json:"fullPluginName,omitempty"`
}

// AdminBlockUserRequest is the AdminBlockUserRequest message type
type AdminBlockUserRequest struct {
	UserID    string `json:"id,omitempty"`
	IsBlocked bool   `json:"blocked,omitempty"`
}

// PickAuthProviderEntryHostOwnerIDType is the PickAuthProviderEntryHostOwnerIDType message type
type PickAuthProviderEntryHostOwnerIDType struct {
	Host string `json:"host,omitempty"`

	// OwnerId string `json:"ownerId,omitempty"`   string is the    string message type
	Type string `json:"type,omitempty"`
}

// PickAuthProviderEntryOwnerID is the PickAuthProviderEntryOwnerID message type
type PickAuthProviderEntryOwnerID struct {
	ID      string `json:"id,omitempty"`
	OwnerID string `json:"ownerId,omitempty"`
}

// PickOAuth2ConfigClientIDClientSecret is the PickOAuth2ConfigClientIDClientSecret message type
type PickOAuth2ConfigClientIDClientSecret struct {
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

// UpdateOwnAuthProviderParams is the UpdateOwnAuthProviderParams message type
type UpdateOwnAuthProviderParams struct {
	Entry interface{} `json:"entry,omitempty"`
}

// CreateApplicationOptions is the CreateApplicationOptions message type
type CreateApplicationOptions struct {
	ContextURL string `json:"contextUrl,omitempty"`
	Mode       string `json:"mode,omitempty"`
}

// Root is the Root message type
type Root map[string]*ResolvedPlugin

// ResolvedPlugin is the ResolvedPlugin message type
type ResolvedPlugin struct {
	FullPluginName string `json:"fullPluginName,omitempty"`
	Kind           string `json:"kind,omitempty"`
	URL            string `json:"url,omitempty"`
}

// UninstallPluginParams is the UninstallPluginParams message type
type UninstallPluginParams struct {
	PluginID string `json:"pluginId,omitempty"`
}

// DeleteOwnAuthProviderParams is the DeleteOwnAuthProviderParams message type
type DeleteOwnAuthProviderParams struct {
	ID string `json:"id,omitempty"`
}

// GuessGitTokenScopesParams is the GuessGitTokenScopesParams message type
type GuessGitTokenScopesParams struct {
	Host         string    `json:"host"`
	RepoURL      string    `json:"repoUrl"`
	GitCommand   string    `json:"gitCommand"`
	CurrentToken *GitToken `json:"currentToken"`
}

type GitToken struct {
	Token  string   `json:"token"`
	User   string   `json:"user"`
	Scopes []string `json:"scopes"`
}

// GuessedGitTokenScopes is the GuessedGitTokenScopes message type
type GuessedGitTokenScopes struct {
	Scopes  []string `json:"scopes,omitempty"`
	Message string   `json:"message,omitempty"`
}

// BrandingLink is the BrandingLink message type
type BrandingLink struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// BrandingSocialLink is the BrandingSocialLink message type
type BrandingSocialLink struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Ide is the Ide message type
type Ide struct {
	HelpMenu         []*BrandingLink `json:"helpMenu,omitempty"`
	Logo             string          `json:"logo,omitempty"`
	ShowReleaseNotes bool            `json:"showReleaseNotes,omitempty"`
}

// Links is the Links message type
type Links struct {
	Footer []*BrandingLink       `json:"footer,omitempty"`
	Header []*BrandingLink       `json:"header,omitempty"`
	Legal  []*BrandingLink       `json:"legal,omitempty"`
	Social []*BrandingSocialLink `json:"social,omitempty"`
}

// Branding is the Branding message type
type Branding struct {
	Favicon  string `json:"favicon,omitempty"`
	Homepage string `json:"homepage,omitempty"`
	Ide      *Ide   `json:"ide,omitempty"`
	Links    *Links `json:"links,omitempty"`

	// Either including domain OR absolute path (interpreted relative to host URL)
	Logo                          string `json:"logo,omitempty"`
	Name                          string `json:"name,omitempty"`
	RedirectURLAfterLogout        string `json:"redirectUrlAfterLogout,omitempty"`
	RedirectURLIfNotAuthenticated string `json:"redirectUrlIfNotAuthenticated,omitempty"`
	ShowProductivityTips          bool   `json:"showProductivityTips,omitempty"`
	StartupLogo                   string `json:"startupLogo,omitempty"`
}

// ApplicationInstanceUser is the ApplicationInstanceUser message type
type ApplicationInstanceUser struct {
	AvatarURL  string `json:"avatarUrl,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	LastSeen   string `json:"lastSeen,omitempty"`
	Name       string `json:"name,omitempty"`
	UserID     string `json:"userId,omitempty"`
}

// SendHeartBeatOptions is the SendHeartBeatOptions message type
type SendHeartBeatOptions struct {
	InstanceID    string  `json:"instanceId,omitempty"`
	RoundTripTime float64 `json:"roundTripTime,omitempty"`
	WasClosed     bool    `json:"wasClosed,omitempty"`
}

// UpdateUserStorageResourceOptions is the UpdateUserStorageResourceOptions message type
type UpdateUserStorageResourceOptions struct {
	Content string `json:"content,omitempty"`
	URI     string `json:"uri,omitempty"`
}

// AdditionalUserData is the AdditionalUserData message type
type AdditionalUserData struct {
	EmailNotificationSettings *EmailNotificationSettings `json:"emailNotificationSettings,omitempty"`
	Platforms                 []*UserPlatform            `json:"platforms,omitempty"`
}

// EmailNotificationSettings is the EmailNotificationSettings message type
type EmailNotificationSettings struct {
	DisallowTransactionalEmails bool `json:"disallowTransactionalEmails,omitempty"`
}

// Identity is the Identity message type
type Identity struct {
	AuthID         string `json:"authId,omitempty"`
	AuthName       string `json:"authName,omitempty"`
	AuthProviderID string `json:"authProviderId,omitempty"`

	// This is a flag that triggers the HARD DELETION of this entity
	Deleted      bool     `json:"deleted,omitempty"`
	PrimaryEmail string   `json:"primaryEmail,omitempty"`
	Readonly     bool     `json:"readonly,omitempty"`
	Tokens       []*Token `json:"tokens,omitempty"`
}

// User is the User message type
type User struct {
	AdditionalData               *AdditionalUserData `json:"additionalData,omitempty"`
	AllowsMarketingCommunication bool                `json:"allowsMarketingCommunication,omitempty"`
	AvatarURL                    string              `json:"avatarUrl,omitempty"`

	// Whether the user has been blocked to use our service, because of TOS violation for example.
	// Optional for backwards compatibility.
	Blocked bool `json:"blocked,omitempty"`

	// The timestamp when the user entry was created
	CreationDate string `json:"creationDate,omitempty"`

	// A map of random settings that alter the behaviour of Bhojpur.NET Platform on a per-user basis
	FeatureFlags *UserFeatureSettings `json:"featureFlags,omitempty"`

	// Optional for backwards compatibility
	FullName string `json:"fullName,omitempty"`

	// The user id
	ID         string      `json:"id,omitempty"`
	Identities []*Identity `json:"identities,omitempty"`

	// Whether the user is logical deleted. This flag is respected by all queries in UserDB
	MarkedDeleted bool   `json:"markedDeleted,omitempty"`
	Name          string `json:"name,omitempty"`

	// whether this user can run applications in privileged mode
	Privileged bool `json:"privileged,omitempty"`

	// The permissions and/or roles the user has
	RolesOrPermissions []string `json:"rolesOrPermissions,omitempty"`
}

// Token is the Token message type
type Token struct {
	ExpiryDate   string   `json:"expiryDate,omitempty"`
	IDToken      string   `json:"idToken,omitempty"`
	RefreshToken string   `json:"refreshToken,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	UpdateDate   string   `json:"updateDate,omitempty"`
	Username     string   `json:"username,omitempty"`
	Value        string   `json:"value,omitempty"`
}

// UserFeatureSettings is the UserFeatureSettings message type
type UserFeatureSettings struct {

	// Permanent feature flags are added to each and every application instance
	// this user starts.
	PermanentWSFeatureFlags []string `json:"permanentWSFeatureFlags,omitempty"`
}

// UserPlatform is the UserPlatform message type
type UserPlatform struct {
	Browser string `json:"browser,omitempty"`

	// Since when does the user have the browser extension installe don this device.
	BrowserExtensionInstalledSince string `json:"browserExtensionInstalledSince,omitempty"`

	// Since when does the user not have the browser extension installed anymore (but previously had).
	BrowserExtensionUninstalledSince string `json:"browserExtensionUninstalledSince,omitempty"`
	FirstUsed                        string `json:"firstUsed,omitempty"`
	LastUsed                         string `json:"lastUsed,omitempty"`
	Os                               string `json:"os,omitempty"`
	UID                              string `json:"uid,omitempty"`
	UserAgent                        string `json:"userAgent,omitempty"`
}

// Requirements is the Requirements message type
type Requirements struct {
	Default     []string `json:"default,omitempty"`
	PrivateRepo []string `json:"privateRepo,omitempty"`
	PublicRepo  []string `json:"publicRepo,omitempty"`
}

// AuthProviderInfo is the AuthProviderInfo message type
type AuthProviderInfo struct {
	AuthProviderID      string        `json:"authProviderId,omitempty"`
	AuthProviderType    string        `json:"authProviderType,omitempty"`
	Description         string        `json:"description,omitempty"`
	DisallowLogin       bool          `json:"disallowLogin,omitempty"`
	HiddenOnDashboard   bool          `json:"hiddenOnDashboard,omitempty"`
	Host                string        `json:"host,omitempty"`
	Icon                string        `json:"icon,omitempty"`
	IsReadonly          bool          `json:"isReadonly,omitempty"`
	LoginContextMatcher string        `json:"loginContextMatcher,omitempty"`
	OwnerID             string        `json:"ownerId,omitempty"`
	Requirements        *Requirements `json:"requirements,omitempty"`
	Scopes              []string      `json:"scopes,omitempty"`
	SettingsURL         string        `json:"settingsUrl,omitempty"`
	Verified            bool          `json:"verified,omitempty"`
}

// GetTokenSearchOptions is the GetTokenSearchOptions message type
type GetTokenSearchOptions struct {
	Host string `json:"host,omitempty"`
}

// SetApplicationTimeoutResult is the SetApplicationTimeoutResult message type
type SetApplicationTimeoutResult struct {
	ResetTimeoutOnApplications []string `json:"resetTimeoutOnApplications,omitempty"`
}

// UserMessage is the UserMessage message type
type UserMessage struct {
	Content string `json:"content,omitempty"`

	// date from where on this message should be shown
	From  string `json:"from,omitempty"`
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}
