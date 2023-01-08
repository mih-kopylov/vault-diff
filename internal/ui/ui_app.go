package ui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/hashicorp/vault/api"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mih-kopylov/vault-diff/vault"
	"github.com/rivo/tview"
	"strconv"
	"strings"
	"time"
)

type selectMode int

const (
	singleSelectMode selectMode = 1 << iota
	multipleSelectMode
)

const (
	selectPageName = "select"
	diffPageName   = "diff"
	helpPageName   = "help"
)

var (
	vaultClient *api.Client
	app         *tview.Application
	pages       *tview.Pages
	selectFlex  *tview.Flex

	leftSelectTree  *tview.TreeView
	rightSelectTree *tview.TreeView

	currentSelectMode       = singleSelectMode
	inOnNodeSelectedHandler = false

	secrets []VaultSecret
)

func RunUiApp(vc *api.Client) error {
	vaultClient = vc
	app = tview.NewApplication()

	secrets = loadSecrets()

	leftSelectTree = createSelectTree(" Select left secret to compare ")
	rightSelectTree = createSelectTree(" Select right secret to compare ")

	selectFlex = tview.NewFlex().
		AddItem(leftSelectTree, 0, 1, true).
		AddItem(rightSelectTree, 0, 1, true)

	pages = tview.NewPages()
	pages.AddAndSwitchToPage(selectPageName, selectFlex, true)

	err := app.SetRoot(pages, true).EnableMouse(true).SetFocus(pages).SetInputCapture(hotKeys).Run()
	if err != nil {
		return err
	}

	return nil
}

func createSelectTree(title string) *tview.TreeView {
	result := tview.NewTreeView()
	createSelectTreeNodes(result)
	result.SetTitle(title)
	result.SetBorder(true)
	result.SetSelectedFunc(onNodeSelectedHandler)
	result.SetChangedFunc(onNodeSelectedHandler)
	return result
}

func createSelectTreeNodes(tree *tview.TreeView) {
	rootNode := tview.NewTreeNode("vault").SetSelectable(false)
	for _, vaultSecret := range secrets {
		sec := vaultSecret
		addNode(rootNode, &sec)
	}

	tree.SetRoot(rootNode)
	tree.SetCurrentNode(rootNode)
}

func onNodeSelectedHandler(node *tview.TreeNode) {
	inOnNodeSelectedHandler = true
	defer func() {
		inOnNodeSelectedHandler = false
	}()
	vaultSecret := getVaultSecretFromNodeReference(node)

	if vaultSecret != nil {
		switch currentSelectMode {
		case singleSelectMode:
			selectNodeWithSecret(leftSelectTree, vaultSecret)
			selectNodeWithSecret(rightSelectTree, vaultSecret)
		case multipleSelectMode:
			break
		default:
			panic(fmt.Sprintf("unsupported select mode: %v", currentSelectMode))
		}
	}
}

func getVaultSecretFromNodeReference(node *tview.TreeNode) *NodeVaultSecret {
	if node == nil {
		return nil
	}
	reference := node.GetReference()
	if reference == nil {
		return nil
	}
	return reference.(*NodeVaultSecret)
}

func selectNodeWithSecret(tree *tview.TreeView, secret *NodeVaultSecret) {
	nodeFound := false
	tree.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		nodeSecret := getVaultSecretFromNodeReference(node)
		if nodeSecret != nil && secret != nil && nodeSecret.Key == secret.Key {
			setCurrentTreeNode(tree, node)
			nodeFound = true
		}
		return true
	})
	if !nodeFound {
		setCurrentTreeNode(tree, nil)
	}
}

func setCurrentTreeNode(tree *tview.TreeView, node *tview.TreeNode) {
	tree.SetCurrentNode(node)
	if !inOnNodeSelectedHandler {
		onNodeSelectedHandler(node)
	}
}

func addNode(node *tview.TreeNode, secret *VaultSecret) {
	if secret.Key.Path != "" {
		for _, part := range strings.Split(strings.Trim(secret.Key.Path, "/"), "/") {
			childFound := false
			for _, child := range node.GetChildren() {
				if child.GetText() == part {
					node = child
					childFound = true
				}
			}
			if !childFound {
				newChild := tview.NewTreeNode(part).SetSelectable(false)
				node.AddChild(newChild)
				node = newChild
			}
		}
	}
	nodeVaultSecret := &NodeVaultSecret{
		VaultSecret:     secret,
		SelectedVersion: secret.Metadata.CurrentVersion,
	}
	title := getSecretTitle(nodeVaultSecret)
	secretNode := tview.NewTreeNode(title).SetSelectable(true).SetReference(nodeVaultSecret)
	node.AddChild(secretNode)
}

func getSecretTitle(secret *NodeVaultSecret) string {
	currentVersionMetadata, ok := secret.Metadata.Versions[strconv.Itoa(secret.SelectedVersion)]
	if !ok {
		return ""
	}
	currentVersionUpdateTime := currentVersionMetadata.CreatedTime
	return fmt.Sprintf("%v:%v (%v)", secret.Key.Key, secret.SelectedVersion,
		currentVersionUpdateTime.Format(time.RFC3339))
}

type VaultSecret struct {
	Key      vault.SecretKey
	Metadata *api.KVMetadata
}

type NodeVaultSecret struct {
	*VaultSecret
	SelectedVersion int
}

func loadSecrets() []VaultSecret {
	allSecrets, err := vault.GetAllSecrets(vaultClient)
	if err != nil {
		panic(err)
	}

	var result []VaultSecret
	for _, secret := range allSecrets {
		metadata, err := vault.ReadSecretMetadata(vaultClient, secret.Path+secret.Key)
		if err != nil {
			panic(err)
		}
		result = append(result, VaultSecret{Key: secret, Metadata: metadata})
	}

	return result
}

func hotKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == 'h' {
		//help is shown in any mode
		createHelpPage()
		return event
	}
	frontPageName, _ := pages.GetFrontPage()
	switch frontPageName {
	case selectPageName:
		return getSelectPageHotKeys(event)
	case diffPageName:
		if event.Key() == tcell.KeyEscape {
			closeDiffPage()
		}
		return event
	case helpPageName:
		if event.Key() == tcell.KeyEscape {
			closeHelpPage()
		}
		return event
	default:
		panic(fmt.Sprintf("unsupported page: '%v'", frontPageName))
	}
}

func getSelectPageHotKeys(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyTab {
		switchLeftRightBoxesFocus()
		return event
	}
	if event.Rune() == 'd' {
		createDiffPage()
		return event
	}
	if event.Rune() == 'r' {
		reloadSecrets()
		return event
	}
	if event.Rune() == 's' {
		setSingleSelectMode()
		return event
	}
	if event.Rune() == 'm' {
		setMultipleSelectMode()
		return event
	}
	if (event.Key() == tcell.KeyRight) && (event.Modifiers()&tcell.ModCtrl > 0) {
		selectNextSecretVersion(func(index int) int {
			return index + 1
		})
		return nil
	}
	if (event.Key() == tcell.KeyLeft) && (event.Modifiers()&tcell.ModCtrl > 0) {
		selectNextSecretVersion(func(index int) int {
			return index - 1
		})
		return nil
	}

	return event
}

func selectNextSecretVersion(handler func(index int) int) {
	currentNode := getCurrentSelectTree().GetCurrentNode()
	nodeVaultSecret := getVaultSecretFromNodeReference(currentNode)
	for currentVersion := handler(nodeVaultSecret.SelectedVersion); ; currentVersion = handler(currentVersion) {
		nextVersion, ok := nodeVaultSecret.Metadata.Versions[strconv.Itoa(currentVersion)]
		if !ok {
			break
		}
		if nextVersion.Destroyed {
			continue
		}
		nodeVaultSecret.SelectedVersion = nextVersion.Version
		currentNode.SetText(getSecretTitle(nodeVaultSecret))
		break
	}
}

func setMultipleSelectMode() {
	currentSelectMode = multipleSelectMode
}

func setSingleSelectMode() {
	currentSelectMode = singleSelectMode
	leftSelectedSecret := getVaultSecretFromNodeReference(leftSelectTree.GetCurrentNode())
	selectNodeWithSecret(rightSelectTree, leftSelectedSecret)
}

func createHelpPage() {
	result := tview.NewTextView()
	result.SetDynamicColors(true).SetTitle("Help").SetBorder(true)
	_, _ = fmt.Fprintf(result, `
[green]===== Select Screen =====[white]
[red]R[white] : refresh secrets metadata from Vault
[red]D[white] : show diff between selected secrets
[red]S[white] : enable "single key" mode - compares different versions of the same secret
[red]M[white] : enable "multiple keys" mode - compares different versions of different secrets
[red]Ctrl + Left[white] : select previous secret version
[red]Ctrl + Right[white] : select next secret version

[green]===== Diff Screen =====[white]
[red]Esc[white] : close diff

[green]===== Help Screen =====[white]
[red]Esc[white] : close help

[green]===== General =====[white]
[red]Ctrl + C[white] : close the application 
`)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(result, 0, 5, true).
			AddItem(nil, 0, 1, false), 0, 3, true).
		AddItem(nil, 0, 1, false)

	pages.AddAndSwitchToPage(helpPageName, flex, true)
}

func closeHelpPage() {
	pages.RemovePage(helpPageName)
}

func reloadSecrets() {
	secrets = loadSecrets()
	createSelectTreeNodes(leftSelectTree)
	createSelectTreeNodes(rightSelectTree)
}

func switchLeftRightBoxesFocus() {
	if getCurrentSelectTree() == leftSelectTree {
		app.SetFocus(rightSelectTree)
	} else {
		app.SetFocus(leftSelectTree)
	}
}

func getCurrentSelectTree() *tview.TreeView {
	if rightSelectTree.HasFocus() {
		return rightSelectTree
	}
	return leftSelectTree
}

func createDiffPage() {
	leftVaultSecret := getVaultSecretFromNodeReference(leftSelectTree.GetCurrentNode())
	rightVaultSecret := getVaultSecretFromNodeReference(rightSelectTree.GetCurrentNode())

	title := fmt.Sprintf(" Difference between %v:%v and %v:%v ",
		leftVaultSecret.Key.String(), leftVaultSecret.SelectedVersion,
		rightVaultSecret.Key.String(), rightVaultSecret.SelectedVersion)

	result := tview.NewTextView()
	result.SetDynamicColors(true).SetTitle(title).SetBorder(true)

	leftSource := fmt.Sprintf("%v:%v", leftVaultSecret.Key.String(), leftVaultSecret.SelectedVersion)
	rightSource := fmt.Sprintf("%v:%v", rightVaultSecret.Key.String(), rightVaultSecret.SelectedVersion)
	leftContent, err := vault.GetSecret(vaultClient, leftVaultSecret.Key.String(), leftVaultSecret.SelectedVersion)
	if err != nil {
		_, _ = fmt.Fprintf(result, "Failed to read secret content")
	}
	rightContent, err := vault.GetSecret(vaultClient, rightVaultSecret.Key.String(), rightVaultSecret.SelectedVersion)
	if err != nil {
		_, _ = fmt.Fprintf(result, "Failed to read secret content")
	}

	edits := myers.ComputeEdits(span.URIFromPath(leftSource), leftContent, rightContent)
	diff := fmt.Sprint(gotextdiff.ToUnified(leftSource, rightSource, leftContent, edits))

	if diff == "" {
		_, _ = fmt.Fprintf(result, "Content is equal")
	} else {
		diffLines := strings.Split(diff, "\n")
		addedColor := "[green]"
		removedColor := "[red]"
		sectionColor := "[cyan]"
		defaultColor := "[white]"
		for i := 0; i < len(diffLines); i++ {
			var col string
			diffLine := diffLines[i]
			if strings.HasPrefix(diffLine, "+") {
				col = addedColor
			} else if strings.HasPrefix(diffLine, "-") {
				col = removedColor
			} else if strings.HasPrefix(diffLine, "@@") {
				col = sectionColor
			} else {
				col = defaultColor
			}
			_, _ = fmt.Fprintln(result, col+diffLine+col)
		}
	}

	pages.AddAndSwitchToPage(helpPageName, result, true)
}

func closeDiffPage() {
	pages.RemovePage(diffPageName)
}
