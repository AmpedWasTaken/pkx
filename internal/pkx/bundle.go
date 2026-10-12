package pkx

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

type Bundle struct {
    Schema int `json:"schema"`
    Name string `json:"name"`
    Packages []BundleGroup `json:"packages"`
}
type BundleGroup struct {
    Manager string `json:"manager"`
    Directory string `json:"directory"`
    Dependencies []string `json:"dependencies"`
    Dev []string `json:"devDependencies,omitempty"`
}

func BundlePlan(root,path string) ([]Step,error) {
    bundlePath,err := filepath.Abs(path)
    if err != nil { return nil,err }
    var bundle Bundle
    if err := readJSON(bundlePath,&bundle); err != nil { return nil,err }
    if bundle.Schema != 1 { return nil,fmt.Errorf("unsupported bundle schema %d (expected 1)",bundle.Schema) }
    if bundle.Name == "" || len(bundle.Packages) == 0 { return nil,fmt.Errorf("bundle must have a name and packages") }
    steps := []Step{}
    initialized := map[string]bool{}
    for i,group := range bundle.Packages {
        dir,err := within(root,group.Directory)
        if err != nil { return nil,fmt.Errorf("packages[%d]: %w",i,err) }
        if len(group.Dependencies)+len(group.Dev) == 0 { return nil,fmt.Errorf("packages[%d]: empty dependencies",i) }
        for _,dep := range append(append([]string{},group.Dependencies...),group.Dev...) {
            if err := validateSpec(dep); err != nil { return nil,fmt.Errorf("packages[%d]: %w",i,err) }
        }
        if group.Manager == "pip" && !initialized[dir] {
            initialized[dir] = true
            venv := filepath.Join(dir,".venv")
            exists,err := readDir(venv)
            if err != nil { return nil,err }
            if !exists { steps = append(steps,Step{dir,"python",pythonCommand(),[]string{"-m","venv",".venv"}}) }
        }
        if len(group.Dependencies)>0 {
            step,err := addStep(dir,group.Manager,group.Dependencies,false)
            if err != nil { return nil,fmt.Errorf("packages[%d]: %w",i,err) }
            steps = append(steps,step)
        }
        if len(group.Dev)>0 {
            step,err := addStep(dir,group.Manager,group.Dev,true)
            if err != nil { return nil,fmt.Errorf("packages[%d]: %w",i,err) }
            steps = append(steps,step)
        }
    }
    return steps,nil
}

func validateSpec(spec string) error {
    if spec == "" || strings.HasPrefix(spec,"-") || strings.ContainsAny(spec,"\n\r\x00\t ") {
        return fmt.Errorf("invalid dependency specifier %q",spec)
    }
    return nil
}

func addStep(dir,manager string, deps []string,dev bool) (Step,error) {
    copied := append([]string{},deps...)
    switch manager {
    case "npm","pnpm","yarn","bun":
        if !fileExists(filepath.Join(dir,"package.json")) { return Step{},fmt.Errorf("%s needs package.json in %s",manager,dir) }
        args := []string{"add"}
        if manager=="npm" { args=[]string{"install"} }
        if dev {
            if manager=="npm" { args=append(args,"--save-dev") } else { args=append(args,"--dev") }
        }
        return Step{dir,manager,manager,append(args,copied...)},nil
    case "composer":
        if !fileExists(filepath.Join(dir,"composer.json")) { return Step{},fmt.Errorf("composer needs composer.json in %s",dir) }
        args:=[]string{"require","--no-interaction"}
        if dev { args=append(args,"--dev") }
        return Step{dir,manager,"composer",append(args,copied...)},nil
    case "pip":
        if dev { return Step{},fmt.Errorf("pip does not support devDependencies in bundle schema 1") }
        venv:=filepath.Join(dir,".venv")
        bin:=filepath.Join(venv,"bin","python")
        if filepath.Separator=='\\' { bin=filepath.Join(venv,"Scripts","python.exe") }
        return Step{dir,manager,bin,append([]string{"-m","pip","install"},copied...)},nil
    default:
        return Step{},fmt.Errorf("bundle manager %q is not supported (npm, pnpm, yarn, bun, composer, pip)",manager)
    }
}

func BundleInit(path string) error {
    f,err:=os.OpenFile(path,os.O_WRONLY|os.O_CREATE|os.O_EXCL,0644)
    if err!=nil { return err }
    defer f.Close()
    _,err=f.Write([]byte("{\n  \"schema\": 1,\n  \"name\": \"my-stack\",\n  \"packages\": [\n    {\"manager\": \"npm\", \"directory\": \".\", \"dependencies\": [\"axios\"]}\n  ]\n}\n"))
    return err
}
