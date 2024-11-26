import os
import pathlib
import subprocess
import sys
import time
import lxml.etree
import lxml.html

DRY_RUN = False

spec = {}

def command(func):
    return func
    # def inner(*args, **kwargs):
    #     if not DRY_RUN:
    #         return func(*args, **kwargs)
    #     else:
    #         print(f"{func.__name__} called with {args} and {kwargs}")
    # return inner()

@command
def ocp_login(server, token, namespace):
    r = subprocess.run(["oc", "login", "--insecure-skip-tls-verify", "--token="+token, "--server="+server], capture_output=True, text=True)
    r = subprocess.run(["oc", "project", namespace], capture_output=True, text=True)
    return 0

@command
def untar(dir: str, file: str):
    os.chdir(dir)
    if file.find('.zip') != -1:
        return subprocess.run(["unzip", file])
    else:
        return subprocess.run(["tar", "xvf", file])

@command
def parseOvf(dir: str, file:str):

    with os.scandir(dir) as entries:
        ovfname = None
        for entry in entries:
            if str(entry).find(".ovf") != -1:
                ovfname = entry.name
                continue

    if ovfname == None:
        return -1

    NAMESPACES = {
        'ovf':  'http://schemas.dmtf.org/ovf/envelope/1',
        'rasd': 'http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData',
        'vssd': 'http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData',
        # 'vmw':  "http://www.vmware.com/schema/ovf",
        # 'xsi': "http://www.w3.org/2001/XMLSchema-instance",
        }
    fullovfpath = dir + "/" + ovfname
    print("Extracting HW informattion from " + ovfname)

    root = lxml.etree.parse(fullovfpath)

    elements = root.xpath('//*/rasd:ElementName/text()', namespaces = NAMESPACES)
    for e in elements:
        if 'CPU' in e:
            spec['vCPUS'] = e.split()[0]
        elif 'memory' in e:
            spec["Memory"] = e.split()[0]

    # Disk information
    all_units = root.xpath('//@ovf:capacityAllocationUnits', namespaces = NAMESPACES)
    # units =  " bytes"
    units = ''
    if len(all_units) > 0:
        match all_units[0][-2:]:
            case '30':
                units =  "Gi"
            case '20':
                units =  "Mi"
            case '10':
                units =  "Ki"
    
    dc = root.xpath('//@ovf:capacity', namespaces = NAMESPACES)
    if len(dc) > 0:
        spec["Disk Capacity"] = f"{dc[0]}{units}"

    oses = root.xpath('//ovf:OperatingSystemSection/ovf:Description/text()', namespaces = NAMESPACES)
    if len(oses) > 0:
        spec['OS'] = oses[0]
    
    # print(spec)
    for k,v in spec.items():
        print(f"\t{k} : {v}")
    return spec

@command
def convert(dir:str, file: str):
    with os.scandir(dir) as entries:
        vmdkname = None
        for entry in entries:
            if str(entry).find(".vmdk") != -1:
                vmdkname = entry.name
                continue
    if vmdkname != None:
        qcow2name = file.replace(".ova", ".qcow2").lower()
        print("Converting {} image to QCOW2".format(vmdkname))
        return subprocess.run(["qemu-img", "convert", "-cpf", "vmdk", "-O", "qcow2", dir + "/" + vmdkname, dir + "/" + qcow2name])
    else:
        return -1

@command
def upload(dir:str, file:str):
    # qcow2file = file.replace(".ova", ".qcow2").lower()
    qcow2file = os.path.splitext(file)[0].lower()
    print(f"Uploading {dir}/{qcow2file}.qcow2")
    os.chdir(dir)
    capacity = "10Gi" if not "Disk Capacity" in spec else spec['Disk Capacity']
    return subprocess.run(["virtctl", "image-upload", "dv", qcow2file, "--force-bind", "--insecure", f"--size={capacity}", f"--image-path={qcow2file}.qcow2"])

@command
def download(url:str):
    os.chdir(os.environ['REFLEX_UPLOADED_FILES_DIR'])
    print("Downloading {} ...".format(url))
    return subprocess.run(["wget", url])

@command
def createBootVolume(dir:str, file:str):

    templatefile = pathlib.Path(os.getenv('APPROOT'), "assets", "datasource-template.yaml")
    with open(templatefile, "r") as yml:
        yaml_content = yml.read()

    pvcname = os.path.splitext(file)[0].lower()
    yaml_content = yaml_content.replace("#NAME#", pvcname)

    os.chdir(dir)
    outfile = pvcname + ".yaml"
    with open(outfile, "w") as output:
        output.write(yaml_content)

    return subprocess.run(["oc", "apply", "-f", outfile])

@command
def cleanup(dir: str, file: str):
    print("Removing temporary files in {} ... ".format(dir))
    try:
        with os.scandir(dir) as entries:
            for entry in entries:
                if entry.is_file():
                    os.unlink(entry.path)
    except Exception as e:
        print(e)
        return -1
    # subprocess.run(["clear"])
    print("Finished importing {}".format(file))
    return True

def main():

    dir = '.' if len(sys.argv) == 1 else sys.argv[1]

    print("Parsing OVF in {} ... ".format(dir))
    r = parseOvf(dir, '')
    if isinstance(r, list):
        for m in r:
            print(m)
    
if __name__ == '__main__':
    main()