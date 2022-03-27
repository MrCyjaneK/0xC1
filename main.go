package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Config struct {
	Branches    []string `json:"branches"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	Platforms   []string `json:"platforms"`
	Repository  string   `json:"repository"`
	Runner      string   `json:"runner"`
	Steps       []string `json:"steps"`
	OK          bool     `json:"ok"`
	JobPlatform string
}

var homepath = "/opt/0xC1"
var outpath = path.Join(homepath, "out")
var workpath = path.Join(homepath, "work")
var configpath = path.Join(homepath, "config")

type Database struct {
	Runners map[string]Runner
	Jobs    map[string]Job
	Archs   []string
}

type Job Config

type Runner struct {
	Name       string
	Dockerfile string
}

var db Database

func init() {
	b, err := ioutil.ReadFile(path.Join(homepath, "db.json"))
	if err != nil {
		db = Database{}
		db.Jobs = make(map[string]Job)
		db.Runners = make(map[string]Runner)
		return
	}
	json.Unmarshal(b, &db)
}

func main() {
	os.MkdirAll(workpath, 0750)
	os.MkdirAll(outpath, 0750)
	refreshConfig()
	refreshRunners()
	refreshJobs()
	saveDb()
	saveHtml()
}

func saveHtml() {
	var output = `<!DOCTYPE html>
<head>
	<title>0xC1 - The one and only CI solution</title>
	<style>
	table {
		font-family: arial, sans-serif;
		border-collapse: collapse;
		width: 100%;
	}
	td, th {
		border: 1px solid #dddddd;
		text-align: left;
		padding: 8px;
	}
	tr:nth-child(even) {
		background-color: #dddddd;
	}
	</style>
</head>
<body>
	<table>
		<tr>
			<th>Name</th>
			<th>Platform</th>
			<th>Status</th>
			<th>Source</th>
			<th>Workdir</th>
			<th>Outdir</th>
		</tr>`
	for key, value := range db.Jobs {
		var successstring = `<td style="color:green;">Success</td>`
		if !value.OK {
			successstring = `<td style="color:red;">Fail</td>`
		}
		output += `<tr>
			<td>` + value.Name + `</td>
			<td>` + value.JobPlatform + `</td>
			` + successstring + `
			<td><a href="` + value.Repository + `" target="_blank">link</a></td>
			<td><a href="work/` + strings.Split(key, "_")[0] + `/` + value.Branches[0] + `/" target="_blank">workdir</a></td>
			<td><a href="out/` + strings.Split(key, "_")[0] + `/` + value.Branches[0] + `/` + value.JobPlatform + `" target="_blank">outdir</a></td>
		</tr>`
	}
	output += `</table>
</body>
</html>`
	ioutil.WriteFile(path.Join(homepath, "index.html"), []byte(output), 0750)
}

func saveDb() {
	data, _ := json.Marshal(db)
	ioutil.WriteFile(path.Join(homepath, "db.json"), data, 0750)
}

func refreshRunners() {
	rp := path.Join(configpath, "runners")
	files, err := ioutil.ReadDir(rp)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		archs_b, err := ioutil.ReadFile(path.Join(rp, f.Name(), "archlist"))
		if err != nil {
			log.Fatal(err)
		}
		archs := strings.Split(strings.ReplaceAll(string(archs_b), "\n", ""), ",")
		log.Println(archs)
		name := f.Name()
		log.Println(name)
		df, err := ioutil.ReadFile(path.Join(rp, f.Name(), "Dockerfile"))
		if err != nil {
			log.Fatal(err)
		}
		db.Runners[name] = Runner{
			Name:       name,
			Dockerfile: string(df),
		}
		for _, arch := range archs {
			log.Println("building image for arch:", arch)
			cmd := exec.Command("docker", "buildx", "build", "--platform", arch, "-t", "0xc1_"+name+":"+strings.ReplaceAll(arch, "/", "_"), ".")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = path.Join(rp, f.Name())
			cmd.Run()
		}
	}
}

func refreshJobs() {
	files, err := ioutil.ReadDir(path.Join(configpath, "jobs"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		name := f.Name()[0 : len(f.Name())-5]
		conf_b, err := ioutil.ReadFile(path.Join(configpath, "jobs", f.Name()))
		if err != nil {
			log.Fatal(err)
		}
		var c Config
		json.Unmarshal(conf_b, &c)
		for _, branch := range c.Branches {
			log.Println(f.Name(), name)
			wp := path.Join(workpath, name, branch)
			op := path.Join(outpath, name, branch)
			_, err = os.Stat(wp)
			if os.IsNotExist(err) {
				cmd := exec.Command("git", "clone", c.Repository, "-b", branch, ".")
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				os.MkdirAll(wp, 0750)
				cmd.Dir = wp
				cmd.Run()
			} else {
				cmd := exec.Command("git", "pull")
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout
				cmd.Dir = wp
				cmd.Run()
			}
			for _, platform := range c.Platforms {
				fop := path.Join(op, platform)
				os.MkdirAll(fop, 0750)
				var buildok = true
				for _, step := range c.Steps {
					cmd := exec.Command("docker", "run", "--rm", "-w", "/opt/work", "--platform", platform, "-v", path.Join(wp)+":/opt/work:ro", "-v", path.Join(fop)+":/opt/out", "0xc1_"+c.Runner+":"+strings.ReplaceAll(platform, "/", "_"), "sh", "-c", step)
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout
					cmd.Dir = wp
					err := cmd.Run()
					if err != nil {
						buildok = false
						break
					}
				}
				c.OK = buildok
				c.JobPlatform = platform
				db.Jobs[name+"_"+platform] = Job(c)
			}
		}
	}
}

func refreshConfig() {
	cmd := exec.Command("git", "pull")
	cmd.Dir = workpath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
