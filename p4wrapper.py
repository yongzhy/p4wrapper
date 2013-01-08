## Place this file into sublimeText 2 installation folder
## under Data\Packages\User
##
## From SublimeText2, Open User Key Binding file From Preferences -> Key Bindings - User, 
## Add following lines:
##  {
##    "command": "p4_checkout",
##    "keys": ["ctrl+alt+e"]
##  },
##  {
##    "command": "p4_revert",
##    "keys": ["ctrl+alt+r"]
##  }
##

import sublime, sublime_plugin
import os

def Checkout(file):
   command = "c:\\portable\\perforce\\p4wrapper.exe"
   command += " -file=" + file 
   command += " -action=edit"
   command += " & pause"
   print command 
   os.system(command)


def Revert(file):
   command = "c:\\portable\\perforce\\p4wrapper.exe"
   command += " -file=" + file 
   command += " -action=revert"
   command += " & pause"
   os.system(command)


class P4CheckoutCommand(sublime_plugin.TextCommand):
   def run(self, edit):
      if(self.view.file_name()):
         Checkout(self.view.file_name())
      else:
         print("View does not contain a file")

class P4RevertCommand(sublime_plugin.TextCommand):
   def run(self, edit):
      if(self.view.file_name()):
         Revert(self.view.file_name())
      else:
         print("View does not contain a file")
