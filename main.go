package main

import (
	//"fmt"
	"html/template"
	"net/http"
	"io"
	"github.com/26tanishabanik/customController/redis"
	"github.com/26tanishabanik/customController/consumer"
	"github.com/26tanishabanik/customController/pods"
	//"github.com/26tanishabanik/customController/deployments"
)

// func main(){
// 	redis.PrintCache()
// }

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("/home/tanisha/Desktop/Projects/customController/templates/*.gohtml"))
}

func main() {
	http.HandleFunc("/", index)
	// http.HandleFunc("/process", process)
	redis.Init()
	go consumer.InitConsumer()
	go pods.Both()
	//go deployments.DeploymentsMain()
	http.ListenAndServe(":9090", nil)

}
func index(w http.ResponseWriter, r *http.Request) {

	tpl.ExecuteTemplate(w, "index.gohtml", nil)
	io.WriteString(w, "Hello, I am Tanisha")
	//io.WriteString(w, vd)
}

// func process(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != "POST" {
// 		http.Redirect(w, r, "/", http.StatusSeeOther)
// 		return
// 	}
// 	vd := redis.PrintCache()
// 	_, err := fmt.Fprintf(w, "Description->%s\n", vd[0])
// 	fmt.Println("Logs struct: ",vd)
// 	fmt.Printf("couldn't write to html page %s", err.Error())
// 	d := struct {
// 					YamlData string
// 				}{
// 					YamlData: vd[0],
// 				}
// 	tpl.ExecuteTemplate(w, "processor.gohtml", d)
// }

// 	contPortint, _ := strconv.ParseInt(contPort, 10, 64)
// 	if kind == "Pod" {
// 		yamlData := Pods.SendYaml(apiversion, kind, topicLabel, chapterLabel, metaname, zone, version, contName, contImage, contPortint)
// 		//yamlData = strings.Replace(yamlData, "\n", `\<br/>\`, -1)
// 		d := struct {
// 			YamlData string
// 		}{
// 			YamlData: yamlData,
// 		}
// 		tpl.ExecuteTemplate(w, "processor.gohtml", d)
// 	}

// }
