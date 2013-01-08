p4wrapper
=========

Simple wrapper for perforce p4 command, only support edit and revert command. Can be integrated with any IDE as external tool. 

I only wrapped edit and revert commands, because I think it is enough for me.

This wrapper will assign P4_CLIENT value automatically according to the absolute path of passed-in file.
The wrapper will first create a file contain all workspaces for current user on current host with detail directory mappings. Later one, this file will be used as lookup database to find out the current working workspace. 

**p4wrapper.go** : The wrapper for p4 command. Need install Go to build

    >go build p4wrapper.go 
    >p4wrapper.exe -file=absolute_path_to_file_to_work_on -action=edit

avariables options for action:

    edit : To check out file from server for edit
    revert : To revert the local copy

Integrate with SublimeText 2
------------------------------

**p4wrapper.py** : The plugin file to be used in SublimeText 2

**Modify the python file for the path to p4warpper.exe**

Place this file into sublimeText 2 installation folder under **Data\Packages\User**

From SublimeText2, Open User Key Binding file From **Preferences -> Key Bindings - User**, 

Add following lines:

    {
      "command": "p4_checkout",
      "keys": ["ctrl+alt+e"]
    },
    {
      "command": "p4_revert",
      "keys": ["ctrl+alt+r"]

Integrate with Source Insight
------------------------
Create customer command for 

    Perforce Edit : path\to\p4wrapper.exe -file=%f -action=edit
    Perforce Revert : path\to\p4wrapper.exe -file=%f -action=edit

Define custome menu items and custome keys for the two new commands.


Integrate with other IDE
-----------------------
Similiar with integrate with SublimeText 2 and Source Insight. Just run this p4wrapper.exe as external tool, pass the absolute path of working file and the action (edit/revert), define your own shortcut keys.


