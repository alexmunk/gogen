'''
Copyright (C) 2005-2012 Splunk Inc. All Rights Reserved.
'''
from __future__ import division
import sys, os
import xml.dom.minidom
import subprocess
import logging, logging.handlers
import platform
import urllib

def setupLogger(logger=None, log_format='%(asctime)s %(levelname)s [Gogen] %(message)s', level=logging.INFO, log_name="gogen.log", logger_name="gogen"):
    """
    Setup a logger suitable for splunkd consumption
    """
    if logger is None:
        logger = logging.getLogger(logger_name)

    logger.propagate = False # Prevent the log messages from being duplicated in the python.log file
    logger.setLevel(level)

    file_handler = logging.handlers.RotatingFileHandler(os.path.join(os.environ['SPLUNK_HOME'], 'var', 'log', 'splunk', log_name), maxBytes=2500000, backupCount=5)
    formatter = logging.Formatter(log_format)
    file_handler.setFormatter(formatter)

    logger.handlers = []
    logger.addHandler(file_handler)

    logger.debug("init %s logger", logger_name)
    return logger

SCHEME = """<scheme>
    <title>Gogen</title>
    <description>Generate data, via Gogen</description>
    <use_external_validation>false</use_external_validation>
    <use_single_instance>false</use_single_instance>
    <streaming_mode>xml</streaming_mode>
    <endpoint/>
</scheme>
"""
def do_scheme():
    print SCHEME


# read XML configuration passed from splunkd
def get_config():
    config = {}

    try:
        # read everything from stdin
        config_str = sys.stdin.read()
        logger.debug("Config Str: %s" % config_str)

        # parse the config XML
        doc = xml.dom.minidom.parseString(config_str)
        root = doc.documentElement
        server_host = str(root.getElementsByTagName("server_host")[0].firstChild.data)
        if server_host:
            logger.debug("XML: Found server_host")
            config["server_host"] = server_host
        server_uri = str(root.getElementsByTagName("server_uri")[0].firstChild.data)
        if server_uri:
            logger.debug("XML: Found server_uri")
            config["server_uri"] = server_uri
        session_key = str(root.getElementsByTagName("session_key")[0].firstChild.data)
        if session_key:
            logger.debug("XML: Found session_key")
            config["session_key"] = session_key
        checkpoint_dir = str(root.getElementsByTagName("checkpoint_dir")[0].firstChild.data)
        if checkpoint_dir:
            logger.debug("XML: Found checkpoint_dir")
            config["checkpoint_dir"] = checkpoint_dir
        conf_node = root.getElementsByTagName("configuration")[0]
        if conf_node:
            logger.debug("XML: found configuration")
            stanza = conf_node.getElementsByTagName("stanza")[0]
            if stanza:
                stanza_name = stanza.getAttribute("name")
                if stanza_name:
                    logger.debug("XML: found stanza " + stanza_name)
                    config["name"] = stanza_name

                    params = stanza.getElementsByTagName("param")
                    for param in params:
                        param_name = param.getAttribute("name")
                        logger.debug("XML: found param '%s'" % param_name)
                        if param_name and param.firstChild and \
                           param.firstChild.nodeType == param.firstChild.TEXT_NODE:
                            data = param.firstChild.data
                            config[param_name] = data
                            logger.debug("XML: '%s' -> '%s'" % (param_name, data))

        checkpnt_node = root.getElementsByTagName("checkpoint_dir")[0]
        if checkpnt_node and checkpnt_node.firstChild and \
           checkpnt_node.firstChild.nodeType == checkpnt_node.firstChild.TEXT_NODE:
            config["checkpoint_dir"] = checkpnt_node.firstChild.data

        if not config:
            raise Exception, "Invalid configuration received from Splunk."

        # just some validation: make sure these keys are present (required)
        # validate_conf(config, "name")
        # validate_conf(config, "key_id")
        # validate_conf(config, "secret_key")
        # validate_conf(config, "checkpoint_dir")
    except Exception, e:
        raise Exception, "Error getting Splunk configuration via STDIN: %s" % str(e)

    return config

if __name__ == '__main__':
    logger = setupLogger(level=logging.DEBUG)

    if len(sys.argv) > 1:
        if sys.argv[1] == "--scheme":
            do_scheme()
            sys.exit(0)
    else:
        config = get_config()

        if platform.system() == 'Linux':
            exefile = 'gogen_real'
            gogen_url = 'http://api.gogen.io/linux/gogen'
        elif platform.system() == 'Windows':
            exefile = 'gogen_real.exe'
            gogen_url = 'http://api.gogen.io/windows/gogen.exe'
        else:
            exefile = 'gogen_real'
            gogen_url = 'http://api.gogen.io/osx/gogen'
        
        gogen_path = os.path.join(os.environ['SPLUNK_HOME'], 'etc', 'apps', 'splunk_app_gogen', 'bin', exefile)
        if not os.path.exists(gogen_path):
            urllib.urlretrieve(gogen_url, gogen_path)
            
        args = [ ]
        args.append(gogen_path)
        args.append('-ot')
        args.append('modinput')

        if 'config' in config:
            args.append('-c')
            args.append(str(config['config']))
        
        args.append('gen')

        if 'count' in config:
            args.append('-c')
            args.append(str(config['count']))
        if 'interval' in config:
            args.append('-i')
            args.append(str(config['interval']))
        if 'endIntervals' in config:
            args.append('-ei')
            args.append(str(config['endIntervals']))
        if 'begin' in config:
            args.append('-b')
            args.append(str(config['begin']))
        if 'end' in config:
            args.append('-e')
            args.append(str(config['end']))
        if 'begin' not in config and 'end' not in config and 'endIntervals' not in config:
            args.append('-r')
        

        import pprint
        logger.debug('args: %s' % pprint.pformat(args))
        logger.debug('command: %s' % ' '.join(args))

        p = subprocess.Popen(args, cwd=os.path.join(os.environ['SPLUNK_HOME'], 'etc', 'apps', 'splunk_app_gogen'),
                            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, shell=False)

        sys.stdout.write("<stream>\n")

        while True:
            data = p.stdout.readline()
            # logger.debug("data: %s" % data)
            sys.stdout.write(data)
