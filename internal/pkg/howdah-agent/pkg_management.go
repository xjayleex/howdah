package howdah_agent

import (
	"context"
	"errors"
	"github.com/contiv/executor"
	"github.com/sirupsen/logrus"
	"howdah/internal/pkg/common/codes"
	"howdah/internal/pkg/common/infra"
	"howdah/internal/pkg/common/status"
	"os/exec"
	"strings"
)



type Packages map[string]Package

func (p Packages) Update(ctx context.Context, cmd *exec.Cmd) error {
	exec := executor.NewCapture(cmd)
	err := exec.Start()
	if err != nil {
		return err
	}
	result, err := exec.Wait(ctx)
	if err != nil {
		return err
	}
	err = p.updateWithExecResult(result)
	return err
}

func (p Packages) updateWithExecResult(result *executor.ExecResult) error {
	parser, err := PackageParserFactory()
	if err != nil {
		return err
	}

	stream := parser.StdoutStream(result)
	for _, line := range stream {
		pkg, err := parser.Parse(line)
		if err != nil {
			// Fragments would be saved in parser cache.
			continue
		}
		if _, exists := p[pkg.name]; !exists {
			p[pkg.name] = pkg
		}
	}

	return nil
}

type Package struct {
	name    string
	version string
	repo    string
}

type PackageParser interface {
	StdoutStream(result *executor.ExecResult) []string
	Parse(line string) (Package, error)
}

func PackageParserFactory() (PackageParser, error) {
	osFamily := infra.GetOsFamily()
	switch osFamily {
	case infra.OsFamily_REDHAT:
		return NewYumPackageParser(), nil
	case infra.OsFamily_DEBIAN:
		return nil, status.Errorf(codes.NotImplemented, "not implemented PackageParserFactory for Debian os family")
	}
	return nil, status.Errorf(codes.UnknownOsFamily, "unknown ")
}

type yumPackageParser struct {
	cache []string
}

func NewYumPackageParser() *yumPackageParser {
	return &yumPackageParser{
		cache: []string{},
	}
}

func (yp *yumPackageParser) StdoutStream(result *executor.ExecResult) []string {
	stream := strings.Split(result.Stdout, "\n")
	for i, line := range stream {
		fields := strings.Fields(line)
		if len(fields) >= 2 && (fields[0] == "Installed" || fields[0] == "Available") && fields[1] == "Packages" {
			stream = stream[i+1:]
			break
		}
	}
	return stream
}

func (yp *yumPackageParser) Parse(line string) (Package, error) {
	fields := strings.Fields(line)
	if len(fields) == 3 {
		pkg := yp.Package(fields[0], fields[1], fields[2])
		return pkg, nil
	}
	if len(fields)+len(yp.cache) == 3 {
		for _, elem := range fields {
			yp.cache = append(yp.cache, elem)
		}
		pkg := yp.Package(yp.cache[0], yp.cache[1], yp.cache[2])
		yp.clearCache()
		return pkg, nil
	} else {
		for _, elem := range fields {
			yp.cache = append(yp.cache, elem)
		}
		return Package{}, errors.New("")
	}
}

func (yp *yumPackageParser) Package(name, version, repo string) Package {
	name = yp.parseName(name)
	repo = yp.parseRepo(repo)
	return Package{
		name:    name,
		version: version,
		repo:    repo,
	}
}

func (yp *yumPackageParser) parseName(name string) string {
	idx := strings.LastIndex(name, ".")
	if idx == -1 {
		return name
	}
	return name[:idx]
}

func (yp *yumPackageParser) parseRepo(repo string) string {
	if repo[0] == '@' {
		return repo[1:]
	} else {
		return repo
	}
}

func (yp *yumPackageParser) clearCache() {
	yp.cache = nil
}

type RepoManager interface {
	InstallPackage(pkgName string, repoFilter ...string) error
}

func NewRepoManager(logger *logrus.Logger) (RepoManager, error) {
	osFamily := infra.GetOsFamily()
	switch osFamily {
	case infra.OsFamily_REDHAT:
		return NewYumRepoManager(logger)
	case infra.OsFamily_DEBIAN:
		// Todo : Return NewAptRepoManager(), nil
		return nil, status.Errorf(codes.NotImplemented, "not implemented RepoManager")
	}
	return nil, status.Errorf(codes.UnknownOsFamily, "unsupported os family")
}

type yumRepoManager struct {
	available       Packages
	installed       Packages
	commandFactory  CommandFactory
	commandExecutor infra.OsCommandExecutor
}

func NewYumRepoManager(logger *logrus.Logger) (*yumRepoManager, error) {
	executor := infra.NewOsCommandExecutor(logger)
	commandFactory := NewRedhatCommandFactory()

	repoManager := &yumRepoManager{
		available:       make(Packages),
		installed:       make(Packages),
		commandFactory:  commandFactory,
		commandExecutor: executor,
	}

	err := repoManager.initialize()
	if err != nil {
		return nil, err
	}
	return repoManager, nil
}

func (rm *yumRepoManager) initialize() error {
	// initialize available
	err := rm.available.Update(context.Background(), rm.commandFactory.ListAvailableCommand())
	err = rm.installed.Update(context.Background(), rm.commandFactory.ListInstalledCommand())

	return err
}

func (rm *yumRepoManager) InstallPackage(pkgName string, repoFilter ...string) error {
	_, installed := rm.Installed(pkgName)
	if installed {
		return status.Errorf(codes.PackageAlreadyInstalled, "")
	}
	pkg, available := rm.Available(pkgName)
	if !available {
		return status.Errorf(codes.PackageNotAvailable, "")
	}

	cmd := rm.commandFactory.InstallCommand(true, pkg)
	if repoFilter != nil {
		if exists := rm.repoFilter(pkg, repoFilter...); exists {
			repoArgs := []string{"--disablerepo=*", "--enablerepo=" + repoFilter[0]}
			cmd = rm.commandFactory.Extend(cmd, repoArgs)
		} else {
			return status.Errorf(codes.PackageNotAvailable, "package %s is not available in repo : %s", pkg.name, pkg.repo)
		}
	}

	// Fixme : Background context would not be adequate.
	_, err := rm.commandExecutor.Execute(context.Background(), cmd)
	if err != nil {
		return status.Errorf(codes.Unknown, "")
	}
	return nil
}

func (rm *yumRepoManager) Available(pkgName string) (pkg Package, exists bool) {
	pkg, exists = rm.available[pkgName]
	return pkg, exists
}

func (rm *yumRepoManager) Installed(pkgName string) (pkg Package, exists bool) {
	pkg, exists = rm.installed[pkgName]
	return pkg, exists
}

func (rm *yumRepoManager) repoFilter(pkg Package, repoFilter ...string) bool {
	if len(repoFilter) != 1 {
		return false
	}
	if strings.Compare(strings.ToLower(pkg.repo), strings.ToLower(repoFilter[0])) != 0 {
		return false
	}

	return true
}


type CommandFactory interface {
	InstallCommand(stdout bool, packages ...Package) *exec.Cmd
	RemoveCommand(stdout bool, packages ...Package) *exec.Cmd
	ListAvailableCommand() *exec.Cmd
	ListInstalledCommand() *exec.Cmd
	ListAllCommand() *exec.Cmd
	Extend(origin *exec.Cmd, additional []string) *exec.Cmd
}

type redhatCommandFactory struct {
	repoManagerBin string
	pkgManagerBin  string

	noStdoutOption []string

	installKeyword []string
	removeKeyword  []string
	listKeyword    []string
}

func NewRedhatCommandFactory() *redhatCommandFactory {
	return &redhatCommandFactory{
		repoManagerBin: "/usr/bin/yum",
		pkgManagerBin:  "/usr/bin/rpm",
		noStdoutOption: []string{"-d", "0", "-e", "0"},
		installKeyword: []string{"-y", "install"},
		removeKeyword:  []string{"-y", "remove"},
		listKeyword:    []string{"list"},
		//listKeyword:    []string{"list", "--showduplicates"},
	}
}

func (r *redhatCommandFactory) InstallCommand(stdout bool, packages ...Package) *exec.Cmd {
	// For example, if we install the "Foo" package through yum,
	// the command line would be :
	// yum -d 0 -e 0 -y install Foo
	// Sine the number of remaining factors except for the yum keyword is
	// generally 7 or more, we set initial capacity to 8.
	args := make([]string, 0, 8)
	if !stdout {
		args = append(args, r.noStdoutOption...)
	}
	args = append(args, r.installKeyword...)
	args = append(args, extractPackageNames(packages...)...)
	cmd := exec.Command(r.repoManagerBin, args...)
	return cmd
}

func (r *redhatCommandFactory) RemoveCommand(stdout bool, packages ...Package) *exec.Cmd {
	args := make([]string, 8)
	if !stdout {
		args = append(args, r.noStdoutOption...)
	}
	args = append(args, r.removeKeyword...)
	args = append(args, extractPackageNames(packages...)...)
	cmd := exec.Command(r.repoManagerBin, args...)
	return cmd
}

func (r *redhatCommandFactory) ListAvailableCommand() *exec.Cmd {
	return r.buildListCommand("available")
}

func (r *redhatCommandFactory) ListInstalledCommand() *exec.Cmd {
	return r.buildListCommand("installed")
}

func (r *redhatCommandFactory) ListAllCommand() *exec.Cmd {
	return r.buildListCommand("all")
}

func (r *redhatCommandFactory) Extend(origin *exec.Cmd, additional []string) *exec.Cmd {
	args := append(origin.Args[1:], additional...)
	extended := exec.Command(origin.Path, args...)
	return extended
}

func (r *redhatCommandFactory) buildListCommand(listOpt string) *exec.Cmd {
	listArgs := r.basicListArgs()
	listArgs = append(listArgs, listOpt)
	cmd := exec.Command(r.repoManagerBin, listArgs...)
	return cmd
}

func (r *redhatCommandFactory) basicListArgs() []string {
	args := make([]string, 0, 4)
	return append(args, r.listKeyword...)
}

func extractPackageNames(packages ...Package) []string {
	var names []string
	for _, pkg := range packages {
		names = append(names, pkg.name)
	}
	return names
}