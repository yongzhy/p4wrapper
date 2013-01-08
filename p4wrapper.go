package main

import (
   "fmt"
   "flag"
   "log"
   "os"
   "os/exec"
   "regexp"
   "strings"
   "io/ioutil"
)

const VERSION = 0.02


var opt_action string
var opt_file   string
var opt_version bool

// Init command arguments.
func init() {
   flag.StringVar(&opt_action, "action", "", "[*] Specify the action to perform (edit, revert)")
   flag.StringVar(&opt_file, "file", "", "[*] Specify the source file to work on")
   flag.BoolVar(&opt_version, "v", false, "Print out version of this software")
   flag.Usage = usage
}

// Print the usage for this utility
func usage() {
   fmt.Printf("%s by Zhu Yong V%0.2f\n", os.Args[0], VERSION)
   fmt.Println("[*] indicate a mandatory command line argument")
   flag.PrintDefaults()
   os.Exit(2)
}

func ParseOptions() {
   // Parse command line arguments.
   flag.Parse()

   if opt_version {
      fmt.Println("Version: ", VERSION)
      os.Exit(0)
   }

   opt_file = strings.TrimSpace(opt_file)
   opt_action = strings.TrimSpace(strings.ToLower(opt_action))

   if opt_file == "" || opt_action == "" {
      flag.Usage()
   }
}

// Get he full path for p4.exe file
func FindP4Cmd() (p4cmd string){
   // Check P4 command line tool is installed or not.
   var err error
   p4cmd, err = exec.LookPath("p4")
   if err != nil {
       log.Fatal("P4 command line is not installed. Please Install it first.")
   }  
   return 
}

// Get username and host name from "p4 set" command
func P4UserAndHost(p4cmd string) (username, hostname string) {
   // Execute "p4 info" command to get username and host information
   out, err := exec.Command(p4cmd, "info").Output()
   if err != nil {
      log.Fatal("\"p4 info\" run error:", err)
   }

   // Get current perforce user name
   re, err := regexp.Compile("User name: ([^\\s]+)")
   ret := re.FindStringSubmatch(string(out))
   if len(ret) < 2 {
      log.Fatal("Can't find user name after \"p4 info\" command")
   }
   username = ret[1]

   // Get current client host name
   re, err = regexp.Compile("Client host: ([^\\s]+)")
   ret = re.FindStringSubmatch(string(out))
   if len(ret) < 2 {
      log.Fatal("Can't find host name after \"p4 info\" command")
   }
   hostname = ret[1]

   return
}

// Check if current user has login perforce server or not.
func P4Login(p4cmd, username string) (ok bool){
   ticket_valid := true
   ok = false

   re, _ := regexp.Compile(username)

   for retry:=0; retry<=3; retry++ {
      out, err := exec.Command(p4cmd, "login", "-s").Output()
      ticket := strings.TrimSpace(string(out))

      if err != nil {
         ticket_valid = false
      } else if ticket == "" || re.FindString(ticket) != username {
         ticket_valid = false
      } else {
         ok = true
         break
      }

      
      if ticket_valid == false && retry<3 {    // No valid login session, require login
         fmt.Println("You need to login Perforce first.")
         cmd := exec.Command(p4cmd, "login")
         cmd.Stdin = os.Stdin
         cmd.Stdout = os.Stdout
         cmd.Stderr = os.Stderr

         err = cmd.Run()
         if err != nil {
            log.Fatal("\"p4 login\" run error:", err)
         }
      }
   }
   return
}

// Get all workspaces on current hostname and below to current user. Store the client
// information into a temp file for later use.
func UpdateClientInfoFile(p4cmd, username, hostname, client_file string) {
   // Get list of workspace by current user
   out, err := exec.Command(p4cmd, "clients", "-u", username).Output()
   if err != nil {
      log.Fatal(err)
   }

   fi, err := os.Create(client_file)
   if err != nil {
      log.Fatal("p4clients.info Open Error:", err)
   }
   defer fi.Close()

   lines := strings.Split(strings.TrimSpace(string(out)), "\n")
   client_re, err := regexp.Compile("Client ([^ ]+) [0-9]+/[0-9]+/[0-9]+ root ([^']+)(.|\\s)+$")
   for _, v := range lines {
      if v != "" {
         ret := client_re.FindStringSubmatch(v)

         if len(ret) >=3 {
            chk_ws := ret[1]
            output, err := exec.Command(p4cmd, "client", "-o", chk_ws).Output()     // run p4 client -o Name to get details.
            if err != nil {
               log.Fatal(output)
            }

            var chk_hostname, chk_root, chk_view string
            valid_client := true
            for _, v := range strings.Split(string(output), "\n") {
               line := strings.TrimSpace(v)
               if line != "" {
                  if strings.Index(line, "Host") == 0 {                            // get workspace HOST name
                     chk_hostname = strings.TrimSpace(strings.Split(line, ":")[1])
                     if chk_hostname != hostname {
                        valid_client = false
                        break
                     }
                  } else if strings.Index(line, "Root") == 0 {                      // get workspace ROOT directory
                     chk_root = strings.Replace(strings.TrimSpace(line[5:]), "\\", "/", -1)
                     if strings.LastIndex(chk_root, "/") != len(chk_root)-1 {
                        chk_root += "/"
                     }
                  } else if strings.Index(line, "//") == 0 {                        // Get the view map list
                     chk_view += strings.Replace(strings.TrimSpace(strings.Split(line, "...")[1]), "//"+chk_ws+"/", chk_root, -1) + ";"
                  }
               }
            }
            if valid_client {
               fi.WriteString(chk_ws + "," + username + "," + chk_hostname + "," + chk_view + "\n")
            }
         }
      }
   }
}


// get list of clients that own by current user and the root for the client
// is same as the root for the file. This is for situation where user has
// multi clients on multi host, and the root is same on different host.
func WorkingWorkspace(p4cmd, username, hostname, file_path string) (workspace string) {
   var client_file string
   if os.IsPathSeparator('\\') {
      client_file = os.TempDir() + "\\" + "p4clients.info"
   } else {
      client_file = os.TempDir() + "/" + "p4clients.info"
   }  

   workspace = FindWorkspaceFromFile(username, hostname, file_path, client_file)
   if workspace == "" {
      fmt.Println("Updateing the client information, please wait...")
      UpdateClientInfoFile(p4cmd, username, hostname, client_file)                  // update client information again.
      workspace =FindWorkspaceFromFile(username, hostname, file_path, client_file)
   }
   return
}

// Function to get the correct workspace from the stored client information file
func FindWorkspaceFromFile(username, hostname, file_path, client_file string) (workspace string) {
   if _, err := os.Stat(client_file); err != nil  && os.IsNotExist(err) {
      return
   }

   fi, err := os.Open(client_file)
   if err != nil {
      log.Fatal(client_file+ " Open Error:", err)
   }
   defer fi.Close()

   data, err := ioutil.ReadAll(fi)

   chk_file := strings.ToLower(strings.Replace(file_path, "\\", "/", -1))
   max_match := 0
   for _, line := range(strings.Split(string(data), "\n")) {
      info := strings.Split(line, ",")
      if len(info) >= 4 {
         chk_ws := info[0]
         chk_user := info[1]
         chk_host := info[2]
         if username == chk_user && chk_host == hostname {
            chk_view := info[3]
            for _, v := range(strings.Split(chk_view, ";")) {
               if path := strings.ToLower(strings.TrimSpace(v)); path != "" {
                  if strings.Index(chk_file, path) == 0 {
                     if len(path) > max_match {
                        workspace = chk_ws 
                        max_match = len(path)
                     }
                  }
               }
            }
         }
      }
   }

   return
}

// Function to call "p4 edit" to checkout file from perforce server
func EditFile(p4cmd, client, file string) {
   out, err := exec.Command(p4cmd, "-c", client, "edit", file).Output()
   if err != nil {
      log.Fatal(err)
   }
   fmt.Println(string(out))
}

// Function to call "p4 revert" to revert changes on file.
func RevertFile(p4cmd, client, file string) {
   out, err := exec.Command(p4cmd, "-c", client, "revert", file).Output()
   if err != nil {
      log.Fatal(err)
   }
   fmt.Println(string(out))
}


func main() {

   ParseOptions()

   p4cmd := FindP4Cmd()
   username, hostname := P4UserAndHost(p4cmd)
   P4Login(p4cmd, username)

   var workspace string 
   if opt_file != "" {
      opt_file = strings.TrimSpace(opt_file)
      workspace = WorkingWorkspace(p4cmd, username, hostname, opt_file)
      if workspace != "" {
         fmt.Println("User:", username, " Host:", hostname, " Workspace:", workspace)
      } else {
         log.Fatal("No valid workspace found for file: ", opt_file)
      }
   }

   if opt_action != "" {
      opt_action = strings.TrimSpace(strings.ToLower(opt_action))
   }

   // Perform action according to command line arguments.
   switch opt_action {
   case "edit":
      EditFile(p4cmd, workspace, opt_file)
   case "revert":
      RevertFile(p4cmd, workspace, opt_file)
   default:
      log.Fatal("Invalid Action")
   }
}