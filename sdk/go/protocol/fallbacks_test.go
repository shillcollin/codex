package protocol

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedRawMessageFallbacksStayContained(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	rawFallbackAllowlist := map[string]struct{}{
		"Account":                                        {},
		"ActivePermissionProfileModification":            {},
		"AppToolsConfig":                                 {},
		"AuthMode":                                       {},
		"ClientRequest":                                  {},
		"CodexErrorInfo":                                 {},
		"CommandAction":                                  {},
		"CommandExecOutputStream":                        {},
		"CommandExecResizeResponse":                      {},
		"CommandExecTerminateResponse":                   {},
		"CommandExecWriteResponse":                       {},
		"ConfigLayerSource":                              {},
		"ConfiguredHookHandler":                          {},
		"ContentItem":                                    {},
		"DynamicToolCallOutputContentItem":               {},
		"ExperimentalFeatureStage":                       {},
		"ExternalAgentConfigImportCompletedNotification": {},
		"ExternalAgentConfigImportResponse":              {},
		"FileSystemPath":                                 {},
		"FileSystemSpecialPath":                          {},
		"FsCopyResponse":                                 {},
		"FsCreateDirectoryResponse":                      {},
		"FsRemoveResponse":                               {},
		"FsUnwatchResponse":                              {},
		"FsWriteFileResponse":                            {},
		"FunctionCallOutputBody":                         {},
		"FunctionCallOutputContentItem":                  {},
		"GuardianApprovalReviewAction":                   {},
		"InputModality":                                  {},
		"LocalShellAction":                               {},
		"LoginAccountParams":                             {},
		"LoginAccountResponse":                           {},
		"LogoutAccountResponse":                          {},
		"McpServerRefreshResponse":                       {},
		"ModelProviderCapabilitiesReadParams":            {},
		"PatchChangeKind":                                {},
		"PermissionProfile":                              {},
		"PermissionProfileFileSystemPermissions":         {},
		"PermissionProfileModificationParams":            {},
		"PermissionProfileSelectionParams":               {},
		"PluginAvailability":                             {},
		"PluginSource":                                   {},
		"PluginShareDeleteResponse":                      {},
		"PluginShareListParams":                          {},
		"PluginUninstallResponse":                        {},
		"ProcessOutputStream":                            {},
		"ReadOnlyAccess":                                 {},
		"ReasoningItemContent":                           {},
		"ReasoningItemReasoningSummary":                  {},
		"ResponseItem":                                   {},
		"ResponsesApiWebSearchAction":                    {},
		"ReviewTarget":                                   {},
		"ServerNotification":                             {},
		"SessionSource":                                  {},
		"SkillsChangedNotification":                      {},
		"SubAgentSource":                                 {},
		"ThreadApproveGuardianDeniedActionResponse":      {},
		"ThreadShellCommandResponse":                     {},
		"ThreadListCwdFilter":                            {},
		"TurnItemsView":                                  {},
		"WebSearchAction":                                {},
	}

	file, err := os.Open(filepath.Join(cwd, "generated.go"))
	if err != nil {
		t.Fatalf("open generated.go: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "type ") || !strings.HasSuffix(line, "= json.RawMessage") {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(line, "type "), " = json.RawMessage")
		if _, ok := rawFallbackAllowlist[name]; !ok {
			t.Fatalf("unexpected new raw-message fallback type: %s", name)
		}
		delete(rawFallbackAllowlist, name)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan generated.go: %v", err)
	}
}
