# coding: utf-8
#!/bin/python
import os

need_build_dir=['src/']
need_build_files=[]

def need_build():
    for file in need_build_files:
        output = os.popen('git show %s'%file)
        show_content = output.read()
        show_content_len = len(show_content)
        if show_content_len > 0:
            return True
        continue
    for dir in need_build_dir: 
        output = os.popen('git show %s'%dir)
        show_content = output.read()
        show_content_len = len(show_content)
        if show_content_len > 0:
            return True
        continue
    return False

if __name__ == "__main__":
    if need_build():
        print("True")
    else:
        print("False")


