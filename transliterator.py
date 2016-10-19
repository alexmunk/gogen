from ConfigParser import ConfigParser
import os
import datetime, time
import sys
import re
import __main__
import logging, logging.handlers
import traceback
import json
import pprint
import copy
import urllib
import types
import random
import csv

# 6/7/14 CS   Adding a new logger adapter class which we will use to override the formatting
#             for all messsages to include the sample they came from
class EventgenAdapter(logging.LoggerAdapter):
    """
    Pass in a sample parameter and prepend sample to all logs
    """
    def process(self, msg, kwargs):
        # Useful for multiprocess debugging to add pid, commented by default
        # return "pid=%s module='%s' sample='%s': %s" % (os.getpid(), self.extra['module'], self.extra['sample'], msg), kwargs
        return "module='%s' sample='%s': %s" % (self.extra['module'], self.extra['sample'], msg), kwargs

    def debugv(self, msg, *args, **kwargs):
        """
        Delegate a debug call to the underlying logger, after adding
        contextual information from this adapter instance.
        """
        msg, kwargs = self.process(msg, kwargs)
        self.logger.debugv(msg, *args, **kwargs)

class Config:
    # Stolen from http://code.activestate.com/recipes/66531/
    # This implements a Borg patterns, similar to Singleton
    # It allows numerous instantiations but always shared state
    __sharedState = {}

    # Internal vars
    _firsttime = True
    _confDict = None

    # Externally used vars
    debug = False
    verbose = False
    splunkEmbedded = False
    sessionKey = None
    grandparentdir = None
    greatgrandparentdir = None
    samples = [ ]
    sampleDir = None
    outputWorkers = None
    generatorWorkers = None
    sampleTimers = [ ]
    __generatorworkers = [ ]
    __outputworkers = [ ]

    # Config file options.  We do not define defaults here, rather we pull them in
    # from eventgen.conf.
    # These are only options which are valid in the 'global' stanza
    # 5/22 CS Except for blacklist, we define that in code, since splunk complains about it in
    # the config files
    threading = None
    disabled = None
    blacklist = ".*\.part"

    __outputPlugins = { }
    __plugins = { }
    outputQueue = None
    generatorQueue = None

    args = None

    disabled = False
    mode = "sample"
    sampletype = "raw"
    interval = 60
    delay = 0
    timeMultiple = 1
    ## 0 means all lines in sample
    count = 0
    ## earliest/latest = now means timestamp replacements default to current time
    earliest = "now"
    latest = "now"
    randomizeEvents = False
    fileMaxBytes = 10485760
    fileBackupFiles = 5
    splunkPort = 8089
    splunkMethod = "https"
    index = "main"
    source = "eventgen"
    sourcetype = "eventgen"
    host = "127.0.0.1"
    generator = "default"
    rater = "config"
    generatorWorkers = 1
    outputWorkers = 1
    timeField = "_raw"

    ## Validations
    _validSettings = ['disabled', 'blacklist', 'spoolDir', 'spoolFile', 'breaker', 'sampletype' , 'interval',
                    'delay', 'count', 'bundlelines', 'earliest', 'latest', 'eai:acl', 'hourOfDayRate',
                    'dayOfWeekRate', 'randomizeCount', 'randomizeEvents', 'outputMode', 'fileName', 'fileMaxBytes',
                    'fileBackupFiles', 'index', 'source', 'sourcetype', 'host', 'hostRegex', 'projectID', 'accessToken',
                    'mode', 'backfill', 'backfillSearch', 'eai:userName', 'eai:appName', 'timeMultiple', 'debug',
                    'minuteOfHourRate', 'timezone', 'dayOfMonthRate', 'monthOfYearRate', 'perDayVolume',
                    'outputWorkers', 'generator', 'rater', 'generatorWorkers', 'timeField', 'sampleDir', 'threading',
                    'profiler', 'maxIntervalsBeforeFlush', 'maxQueueLength',
                    'verbose', 'useOutputQueue', 'seed','end', 'autotimestamps', 'autotimestamp']
    _validTokenTypes = {'token': 0, 'replacementType': 1, 'replacement': 2}
    _validHostTokens = {'token': 0, 'replacement': 1}
    _validReplacementTypes = ['static', 'timestamp', 'replaytimestamp', 'random', 'rated', 'file', 'mvfile', 'integerid']
    _validOutputModes = [ ]
    _intSettings = ['interval', 'outputWorkers', 'generatorWorkers', 'maxIntervalsBeforeFlush', 'maxQueueLength']
    _floatSettings = ['randomizeCount', 'delay', 'timeMultiple']
    _boolSettings = ['disabled', 'randomizeEvents', 'bundlelines', 'profiler', 'useOutputQueue', 'autotimestamp']
    _jsonSettings = ['hourOfDayRate', 'dayOfWeekRate', 'minuteOfHourRate', 'dayOfMonthRate', 'monthOfYearRate', 'autotimestamps']
    _defaultableSettings = ['disabled', 'spoolDir', 'spoolFile', 'breaker', 'sampletype', 'interval', 'delay',
                            'count', 'bundlelines', 'earliest', 'latest', 'hourOfDayRate', 'dayOfWeekRate',
                            'randomizeCount', 'randomizeEvents', 'outputMode', 'fileMaxBytes', 'fileBackupFiles',
                            'splunkHost', 'splunkPort', 'splunkMethod', 'index', 'source', 'sourcetype', 'host', 'hostRegex',
                            'projectID', 'accessToken', 'mode', 'minuteOfHourRate', 'timeMultiple', 'dayOfMonthRate',
                            'monthOfYearRate', 'perDayVolume', 'sessionKey', 'generator', 'rater', 'timeField', 'maxQueueLength',
                            'maxIntervalsBeforeFlush', 'autotimestamp']
    _complexSettings = { 'sampletype': ['raw', 'csv'],
                         'mode': ['sample', 'replay'],
                         'threading': ['thread', 'process']}

    def __init__(self, args=None):
        """Setup Config object.  Sets up Logging and path related variables."""
        # Rebind the internal datastore of the class to an Instance variable
        self.__dict__ = self.__sharedState
        if self._firsttime:
            # 2/1/15 CS  Adding support for command line arguments
            if args:
                self.args = args

            # Setup logger
            # 12/8/13 CS Adding new verbose log level to make this a big more manageable
            DEBUG_LEVELV_NUM = 9
            logging.addLevelName(DEBUG_LEVELV_NUM, "DEBUGV")
            logging.__dict__['DEBUGV'] = DEBUG_LEVELV_NUM
            def debugv(self, message, *args, **kws):
                # Yes, logger takes its '*args' as 'args'.
                if self.isEnabledFor(DEBUG_LEVELV_NUM):
                    self._log(DEBUG_LEVELV_NUM, message, args, **kws)
            logging.Logger.debugv = debugv

            logger = logging.getLogger('eventgen')
            logger.propagate = False # Prevent the log messages from being duplicated in the python.log file
            logger.setLevel(logging.INFO)
            formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s')
            streamHandler = logging.StreamHandler(sys.stderr)
            streamHandler.setFormatter(formatter)
            # 2/1/15 CS  Adding support for command line arguments.  In this case, if we're running from the command
            # line and we have arguments, we only want output from logger if we're in verbose
            if self.args:
                if self.args.verbosity >= 1:
                    logger.addHandler(streamHandler)
                else:
                    logger.addHandler(logging.NullHandler())

            else:
                logger.addHandler(streamHandler)
            # logging.disable(logging.INFO)

            adapter = EventgenAdapter(logger, {'sample': 'null', 'module': 'config'})
            # Having logger as a global is just damned convenient
            self.logger = adapter

            # Determine some path names in our environment
            self.grandparentdir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
            self.greatgrandparentdir = os.path.dirname(self.grandparentdir)


            self._complexSettings['timezone'] = self._validateTimezone

            self._complexSettings['count'] = self._validateCount

            self._complexSettings['seed'] = self._validateSeed

            self._firsttime = False
            self.intervalsSinceFlush = { }

    def __str__(self):
        """Only used for debugging, outputs a pretty printed representation of our Config"""
        # Filter items from config we don't want to pretty print
        filter_list = [ 'samples', 'sampleTimers', '__generatorworkers', '__outputworkers' ]
        # Eliminate recursive going back to parent
        temp = dict([ (key, value) for (key, value) in self.__dict__.items() if key not in filter_list ])

        return 'Config:'+pprint.pformat(temp)+'\nSamples:\n'+pprint.pformat(self.samples)

    def __repr__(self):
        return self.__str__()

    def parse(self):
        """Parse configs from Splunk REST Handler or from files.
        We get called manually instead of in __init__ because we need find out if we're Splunk embedded before
        we figure out how to configure ourselves.
        """
        self.logger.debug("Parsing configuration files.")
        self._buildConfDict()
        # Set defaults config instance variables to 'global' section
        # This establishes defaults for other stanza settings
        # for key, value in self._confDict['global'].items():
        #     value = self._validateSetting('global', key, value)
        #     setattr(self, key, value)

        # del self._confDict['global']
        if 'default' in self._confDict:
            del self._confDict['default']

        tempsamples = [ ]
        tempsamples2 = [ ]

        # 1/16/16 CS Trying to clean up the need to have attributes hard coded into the Config object
        # and instead go off the list of valid settings that could be set
        for setting in self._validSettings:
            if not hasattr(self, setting):
                setattr(self, setting, None)

        # Now iterate for the rest of the samples we've found
        # We'll create Sample objects for each of them
        for stanza, settings in self._confDict.items():
            sampleexists = False
            for sample in self.samples:
                if sample.name == stanza:
                    sampleexists = True

            # If we see the sample in two places, use the first and ignore the second
            if not sampleexists:
                s = Sample(stanza)
                for key, value in settings.items():
                    oldvalue = value
                    try:
                        value = self._validateSetting(stanza, key, value)
                    except ValueError:
                        # If we're improperly formatted, skip to the next item
                        continue
                    # If we're a tuple, then this must be a token
                    if type(value) == tuple:
                        # Token indices could be out of order, so we must check to
                        # see whether we have enough items in the list to update the token
                        # In general this will keep growing the list by whatever length we need
                        if(key.find("host.") > -1):
                            # self.logger.info("hostToken.{} = {}".format(value[1],oldvalue))
                            if not isinstance(s.hostToken, Token):
                                s.hostToken = Token(s)
                                # default hard-coded for host replacement
                                s.hostToken.replacementType = 'file'
                            setattr(s.hostToken, value[0], oldvalue)
                        else:
                            if len(s.tokens) <= value[0]:
                                x = (value[0]+1) - len(s.tokens)
                                s.tokens.extend([None for i in xrange(0, x)])
                            if not isinstance(s.tokens[value[0]], Token):
                                s.tokens[value[0]] = Token(s)
                            # logger.info("token[{}].{} = {}".format(value[0],value[1],oldvalue))
                            setattr(s.tokens[value[0]], value[1], oldvalue)
                    elif key == 'eai:acl':
                        setattr(s, 'app', value['app'])
                    else:
                        setattr(s, key, value)
                        # 6/22/12 CS Need a way to show a setting was set by the original
                        # config read
                        s._lockedSettings.append(key)
                        # self.logger.debug("Appending '%s' to locked settings for sample '%s'" % (key, s.name))


                # Validate all the tokens are fully setup, can't do this in _validateSettings
                # because they come over multiple lines
                # Don't error out at this point, just log it and remove the token and move on
                deleteidx = [ ]
                for i in xrange(0, len(s.tokens)):
                    t = s.tokens[i]
                    # If the index doesn't exist at all
                    if t == None:
                        self.logger.info("Token at index %s invalid" % i)
                        # Can't modify list in place while we're looping through it
                        # so create a list to remove later
                        deleteidx.append(i)
                    elif t.token == None or t.replacementType == None or t.replacement == None:
                        self.logger.info("Token at index %s invalid" % i)
                        deleteidx.append(i)
                newtokens = [ ]
                for i in xrange(0, len(s.tokens)):
                    if i not in deleteidx:
                        newtokens.append(s.tokens[i])
                s.tokens = newtokens

                # Must have eai:acl key to determine app name which determines where actual files are
                if s.app == None:
                    self.logger.error("App not set for sample '%s' in stanza '%s'" % (s.name, stanza))
                    raise ValueError("App not set for sample '%s' in stanza '%s'" % (s.name, stanza))

                # Set defaults for items not included in the config file
                for setting in self._defaultableSettings:
                    if not hasattr(s, setting) or getattr(s, setting) == None:
                        try:
                            setattr(s, setting, getattr(self, setting))
                        except AttributeError, e:
                            pass

                # Append to temporary holding list
                if not s.disabled:
                    s._priority = len(tempsamples)+1
                    tempsamples.append(s)

        # 6/22/12 CS Rewriting the config matching code yet again to handling flattening better.
        # In this case, we're now going to match all the files first, create a sample for each of them
        # and then take the match from the sample seen last in the config file, and apply settings from
        # every other match to that one.
        for s in tempsamples:
            # Now we need to match this up to real files.  May generate multiple copies of the sample.
            foundFiles = [ ]

            # 1/5/14 Adding a config setting to override sample directory, primarily so I can put tests in their own
            # directories
            if s.sampleDir == None:
                self.logger.debug("Sample directory not specified in config, setting based on standard")
                if self.splunkEmbedded and not STANDALONE:
                    s.sampleDir = os.path.join(self.greatgrandparentdir, s.app, 'samples')
                else:
                    # 2/1/15 CS  Adding support for looking for samples based on the config file specified on
                    # the command line.
                    if self.args:
                        if os.path.isdir(self.args.configfile):
                            self.logger.debug("Configfile specified: %s", self.args.configfile)
                            s.sampleDir = os.path.join(self.args.configfile, 'samples')
                        else:
                            s.sampleDir = os.path.join(os.getcwd(), 'samples')
                    else:
                        s.sampleDir = os.path.join(os.getcwd(), 'samples')
                    if not os.path.exists(s.sampleDir):
                        newSampleDir = os.path.join(os.sep.join(os.getcwd().split(os.sep)[:-1]), 'samples')
                        self.logger.error("Path not found for samples '%s', trying '%s'" % (s.sampleDir, newSampleDir))
                        s.sampleDir = newSampleDir

                        if not os.path.exists(s.sampleDir):
                            newSampleDir = os.path.join(self.grandparentdir, 'samples')
                            self.logger.error("Path not found for samples '%s', trying '%s'" % (s.sampleDir, newSampleDir))
                            s.sampleDir = newSampleDir
            else:
                self.logger.debug("Sample directory specified in config, checking for relative")
                # Allow for relative paths to the base directory
                if not os.path.exists(s.sampleDir):
                    s.sampleDir = os.path.join(self.grandparentdir, s.sampleDir)
                else:
                    s.sampleDir = s.sampleDir


            if os.path.exists(s.sampleDir):
                sampleFiles = os.listdir(s.sampleDir)
                for sample in sampleFiles:
                    results = re.match(s.name, sample)
                    if results != None:
                        samplePath = os.path.join(s.sampleDir, sample)
                        if os.path.isfile(samplePath):
                            self.logger.debug("Found sample file '%s' for app '%s' using config '%s' with priority '%s'; adding to list" \
                                % (sample, s.app, s.name, s._priority) )
                            foundFiles.append(samplePath)
            # If we didn't find any files, log about it
            if len(foundFiles) == 0:
                self.logger.warning("Sample '%s' in config but no matching files" % s.name)
                # 1/23/14 Change in behavior, go ahead and add the sample even if we don't find a file
                # 9/16/15 Change bit us, now only append if we're a generator other than the two stock generators
                if not s.disabled and not (s.generator == "default" or s.generator == "replay"):
                    tempsamples2.append(copy.deepcopy(s))
            for f in foundFiles:
                news = copy.deepcopy(s)
                news.filePath = f
                # 12/3/13 CS TODO These are hard coded but should be handled via the modular config system
                # Maybe a generic callback for all plugins which will modify sample based on the filename
                # found?
                # Override <SAMPLE> with real name
                if s.outputMode == 'spool' and s.spoolFile == self.spoolFile:
                    news.spoolFile = f.split(os.sep)[-1]
                if s.outputMode == 'file' and s.fileName == None and s.spoolFile == self.spoolFile:
                    news.fileName = os.path.join(s.spoolDir, f.split(os.sep)[-1])
                elif s.outputMode == 'file' and s.fileName == None and s.spoolFile != None:
                    news.fileName = os.path.join(s.spoolDir, s.spoolFile)
                # Override s.name with file name.  Usually they'll match unless we've been a regex
                # 6/22/12 CS Save original name for later matching
                news._origName = news.name
                news.name = f.split(os.sep)[-1]
                if not news.disabled:
                    tempsamples2.append(news)
                else:
                    self.logger.info("Sample '%s' for app '%s' is marked disabled." % (news.name, news.app))

        # Clear tempsamples, we're going to reuse it
        tempsamples = [ ]

        # We're now going go through the samples and attempt to apply any matches from other stanzas
        # This allows us to specify a wildcard at the beginning of the file and get more specific as we go on

        # Loop through all samples, create a list of the master samples
        for s in tempsamples2:
            foundHigherPriority = False
            othermatches = [ ]
            # If we're an exact match, don't go looking for higher priorities
            if not s.name == s._origName:
                for matchs in tempsamples2:
                    if matchs.filePath == s.filePath and s._origName != matchs._origName:
                        # We have a match, now determine if we're higher priority or not
                            # If this is a longer pattern or our match is an exact match
                            # then we're a higher priority match
                        if len(matchs._origName) > len(s._origName) or matchs.name == matchs._origName:
                            # if s._priority < matchs._priority:
                            self.logger.debug("Found higher priority for sample '%s' with priority '%s' from sample '%s' with priority '%s'" \
                                        % (s._origName, s._priority, matchs._origName, matchs._priority))
                            foundHigherPriority = True
                            break
                        else:
                            othermatches.append(matchs._origName)
            if not foundHigherPriority:
                self.logger.debug("Chose sample '%s' from samples '%s' for file '%s'" \
                            % (s._origName, othermatches, s.name))
                tempsamples.append(s)

        # Now we have two lists, tempsamples which contains only the highest priority matches, and
        # tempsamples2 which contains all matches.  We need to now flatten the config in order to
        # take all the configs which might match.

        # Reversing tempsamples2 in order to look from the bottom of the file towards the top
        # We want entries lower in the file to override entries higher in the file

        tempsamples2.reverse()

        # Loop through all samples
        for s in tempsamples:
            # Now loop through the samples we've matched with files to see if we apply to any of them
            for overridesample in tempsamples2:
                if s.filePath == overridesample.filePath and s._origName != overridesample._origName:
                    # Now we're going to loop through all valid settings and set them assuming
                    # the more specific object that we've matched doesn't already have them set
                    for settingname in self._validSettings:
                        if settingname not in ['eai:acl', 'blacklist', 'disabled', 'name']:
                            # 7/16/14 CS For some reason default settings are suddenly erroring
                            # not sure why, but lets just move on
                            try:
                                sourcesetting = getattr(overridesample, settingname)
                                destsetting = getattr(s, settingname)
                                # We want to check that the setting we're copying to hasn't been
                                # set, otherwise keep the more specific value
                                # 6/22/12 CS Added support for non-overrideable (locked) settings
                                # logger.debug("Locked settings: %s" % pprint.pformat(matchs._lockedSettings))
                                # if settingname in matchs._lockedSettings:
                                #     logger.debug("Matched setting '%s' in sample '%s' lockedSettings" \
                                #         % (settingname, matchs.name))
                                if (destsetting == None or destsetting == getattr(self, settingname)) \
                                        and sourcesetting != None and sourcesetting != getattr(self, settingname) \
                                        and not settingname in s._lockedSettings:
                                    self.logger.debug("Overriding setting '%s' with value '%s' from sample '%s' to sample '%s' in app '%s'" \
                                                    % (settingname, sourcesetting, overridesample._origName, s.name, s.app))
                                    setattr(s, settingname, sourcesetting)
                            except AttributeError:
                                pass

                    # Now prepend all the tokens to the beginning of the list so they'll be sure to match first
                    newtokens = copy.deepcopy(s.tokens)
                    # self.logger.debug("Prepending tokens from sample '%s' to sample '%s' in app '%s': %s" \
                    #             % (overridesample._origName, s.name, s.app, pprint.pformat(newtokens)))
                    newtokens.extend(copy.deepcopy(overridesample.tokens))
                    s.tokens = newtokens

        # We've added replay mode, so lets loop through the samples again and set the earliest and latest
        # settings for any samples that were set to replay mode
        for s in tempsamples:
            # We've added replay mode, so lets loop through the samples again and set the earliest and latest
            # settings for any samples that were set to replay mode
            if s.perDayVolume:
                self.logger.info("Stanza contains per day volume, changing rater and generator to perdayvolume instead of count")
                s.rater = 'perdayvolume'
                s.count = 1
                s.generator = 'perdayvolumegenerator'

            if s.mode == 'replay':
                self.logger.debug("Setting defaults for replay samples")
                s.earliest = 'now'
                s.latest = 'now'
                s.count = 1
                s.randomizeCount = None
                s.hourOfDayRate = None
                s.dayOfWeekRate = None
                s.minuteOfHourRate = None
                s.interval = 0
                # 12/29/13 CS Moved replay generation to a new replay generator plugin
                s.generator = 'replay'

        self.samples = tempsamples
        self._confDict = None

        # 9/2/15 Try autotimestamp values, add a timestamp if we find one
        for s in self.samples:
            self.logger.debug("Generator '%s' for sample '%s'" % (s.generator, s.name))
            if s.generator in ('default', 'replay'):
                s.loadSample()

                if s.autotimestamp:
                    at = self.autotimestamps
                    line_puncts = [ ]

                    # Check for _time field, if it exists, add a timestamp to support it
                    if len(s.sampleDict) > 0:
                        if '_time' in s.sampleDict[0]:
                            self.logger.debugv("Found _time field, checking if default timestamp exists")
                            t = Token()
                            t.token = "\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}"
                            t.replacementType = "timestamp"
                            t.replacement = "%Y-%m-%dT%H:%M:%S.%f"

                            found_token = False
                            # Check to see if we're already a token
                            for st in s.tokens:
                                if st.token == t.token and st.replacement == t.replacement:
                                    found_token = True
                                    break
                            if not found_token:
                                self.logger.debugv("Found _time adding timestamp to support")
                                s.tokens.append(t)
                            else:
                                self.logger.debugv("_time field exists and timestamp already configured")

                    for e in s.sampleDict:
                        # Run punct against the line, make sure we haven't seen this same pattern
                        # Not totally exact but good enough for Rock'N'Roll
                        p = self._punct(e['_raw'])
                        # self.logger.debugv("Got punct of '%s' for event '%s'" % (p, e[s.timeField]))
                        if p not in line_puncts:
                            for x in at:
                                t = Token()
                                t.token = x[0]
                                t.replacementType = "timestamp"
                                t.replacement = x[1]

                                try:
                                    # self.logger.debugv("Trying regex '%s' for format '%s' on '%s'" % (x[0], x[1], e[s.timeField]))
                                    ts = s.getTSFromEvent(e['_raw'], t)
                                    if type(ts) == datetime.datetime:
                                        found_token = False
                                        # Check to see if we're already a token
                                        for st in s.tokens:
                                            if st.token == t.token and st.replacement == t.replacement:
                                                found_token = True
                                                break
                                        if not found_token:
                                            self.logger.debugv("Found timestamp '%s', extending token with format '%s'" % (x[0], x[1]))
                                            s.tokens.append(t)
                                            # Drop this pattern from ones we try in the future
                                            at = [ z for z in at if z[0] != x[0] ]
                                        break
                                except ValueError:
                                    pass
                        line_puncts.append(p)



        self.logger.debug("Finished parsing.  Config str:\n%s" % self)

    def _validateSetting(self, stanza, key, value):
        """Validates settings to ensure they won't cause errors further down the line.
        Returns a parsed value (if the value is something other than a string).
        If we've read a token, which is a complex config, returns a tuple of parsed values."""
        self.logger.debugv("Validating setting for '%s' with value '%s' in stanza '%s'" % (key, value, stanza))
        if key.find('token.') > -1:
            results = re.match('token\.(\d+)\.(\w+)', key)
            if results != None:
                groups = results.groups()
                if groups[1] not in self._validTokenTypes:
                    self.logger.error("Could not parse token index '%s' token type '%s' in stanza '%s'" % \
                                    (groups[0], groups[1], stanza))
                    raise ValueError("Could not parse token index '%s' token type '%s' in stanza '%s'" % \
                                    (groups[0], groups[1], stanza))
                if groups[1] == 'replacementType':
                    if value not in self._validReplacementTypes:
                        self.logger.error("Invalid replacementType '%s' for token index '%s' in stanza '%s'" % \
                                    (value, groups[0], stanza))
                        raise ValueError("Could not parse token index '%s' token type '%s' in stanza '%s'" % \
                                    (groups[0], groups[1], stanza))
                return (int(groups[0]), groups[1])
        elif key.find('host.') > -1:
            results = re.match('host\.(\w+)', key)
            if results != None:
                groups = results.groups()
                if groups[0] not in self._validHostTokens:
                    self.logger.error("Could not parse host token type '%s' in stanza '%s'" % (groups[0], stanza))
                    raise ValueError("Could not parse host token type '%s' in stanza '%s'" % (groups[0], stanza))
                return (groups[0], value)
        elif key in self._validSettings:
            if key in self._intSettings:
                try:
                    value = int(value)
                except:
                    self.logger.error("Could not parse int for '%s' in stanza '%s'" % (key, stanza))
                    raise ValueError("Could not parse int for '%s' in stanza '%s'" % (key, stanza))
            elif key in self._floatSettings:
                try:
                    value = float(value)
                except:
                    self.logger.error("Could not parse float for '%s' in stanza '%s'" % (key, stanza))
                    raise ValueError("Could not parse float for '%s' in stanza '%s'" % (key, stanza))
            elif key in self._boolSettings:
                try:
                    # Splunk gives these to us as a string '0' which bool thinks is True
                    # ConfigParser gives 'false', so adding more strings
                    if value in ('0', 'false', 'False'):
                        value = 0
                    value = bool(value)
                except:
                    self.logger.error("Could not parse bool for '%s' in stanza '%s'" % (key, stanza))
                    raise ValueError("Could not parse bool for '%s' in stanza '%s'" % (key, stanza))
            elif key in self._jsonSettings:
                try:
                    value = json.loads(value)
                except:
                    self.logger.error("Could not parse json for '%s' in stanza '%s'" % (key, stanza))
                    raise ValueError("Could not parse json for '%s' in stanza '%s'" % (key, stanza))
            # 12/3/13 CS Adding complex settings, which is a dictionary with the key containing
            # the config item name and the value is a list of valid values or a callback function
            # which will parse the value or raise a ValueError if it is unparseable
            elif key in self._complexSettings:
                complexSetting = self._complexSettings[key]
                self.logger.debugv("Complex setting for '%s' in stanza '%s'" % (key, stanza))
                # Set value to result of callback, e.g. parsed, or the function should raise an error
                if isinstance(complexSetting, types.FunctionType) or isinstance(complexSetting, types.MethodType):
                    self.logger.debugv("Calling function for setting '%s' with value '%s'" % (key, value))
                    value = complexSetting(value)
                elif isinstance(complexSetting, list):
                    if not value in complexSetting:
                        self.logger.error("Setting '%s' is invalid for value '%s' in stanza '%s'" % (key, value, stanza))
                        raise ValueError("Setting '%s' is invalid for value '%s' in stanza '%s'" % (key, value, stanza))
        else:
            # Notifying only if the setting isn't valid and continuing on
            # This will allow future settings to be added and be backwards compatible
            self.logger.warning("Key '%s' in stanza '%s' is not a valid setting" % (key, stanza))
        return value

    def _validateTimezone(self, value):
        """Callback for complexSetting timezone which will parse and validate the timezone"""
        self.logger.debug("Parsing timezone '%s'" % (value))
        if value.find('local') >= 0:
            value = datetime.timedelta(days=1)
        else:
            try:
                # Separate the hours and minutes (note: minutes = the int value - the hour portion)
                if int(value) > 0:
                    mod = 100
                else:
                    mod = -100
                value = datetime.timedelta(hours=int(int(value) / 100.0), minutes=int(value) % mod )
            except:
                self.logger.error("Could not parse timezone '%s' for '%s'" % (value, key))
                raise ValueError("Could not parse timezone '%s' for '%s'" % (value, key))
        self.logger.debug("Parsed timezone '%s'" % (value))
        return value

    def _validateCount(self, value):
        """Callback to override count to -1 if set to 0 in the config, otherwise return int"""
        self.logger.debug("Validating count of %s" % value)
        # 5/13/14 CS Hack to take a zero count in the config and set it to a value which signifies
        # the special condition rather than simply being zero events, setting to -1
        try:
            value = int(value)
        except:
            self.logger.error("Could not parse int for 'count' in stanza '%s'" % (key, stanza))
            raise ValueError("Could not parse int for 'count' in stanza '%s'" % (key, stanza))

        if value == 0:
            value = -1
        self.logger.debug("Count set to %d" % value)

        return value

    def _validateSeed(self, value):
        """Callback to set random seed"""
        self.logger.debug("Validating random seed of %s" % value)
        try:
            value = int(value)
        except:
            self.logger.error("Could not parse int for 'seed' in stanza '%s'" % (key, stanza))
            raise ValueError("Could not parse int for 'count' in stanza '%s'" % (key, stanza))

        self.logger.info("Using random seed %s" % value)
        random.seed(value)



    def _buildConfDict(self):
        """Build configuration dictionary that we will use """

        # Abstracts grabbing configuration from Splunk or directly from Configuration Files

        if self.splunkEmbedded and not STANDALONE:
            self.logger.info('Retrieving eventgen configurations from /configs/eventgen')
            self._confDict = entity.getEntities('configs/eventgen', count=-1, sessionKey=self.sessionKey)
        else:
            self.logger.info('Retrieving eventgen configurations with ConfigParser()')
            # We assume we're in a bin directory and that there are default and local directories
            conf = ConfigParser()
            # Make case sensitive
            conf.optionxform = str
            currentdir = os.getcwd()

            conffiles = [ ]
            # 2/1/15 CS  Moving to argparse way of grabbing command line parameters
            if self.args:
                if self.args.configfile:
                    if os.path.exists(self.args.configfile):
                        # 2/1/15 CS Adding a check to see whether we're instead passed a directory
                        # In which case we'll assume it's a splunk app and look for config files in
                        # default and local
                        if os.path.isdir(self.args.configfile):
                            conffiles = [os.path.join(self.grandparentdir, 'default', 'eventgen.conf'),
                                    os.path.join(self.args.configfile, 'default', 'eventgen.conf'),
                                    os.path.join(self.args.configfile, 'local', 'eventgen.conf')]
                        else:
                            conffiles = [os.path.join(self.grandparentdir, 'default', 'eventgen.conf'),
                                    self.args.configfile]
            if len(conffiles) == 0:
                conffiles = [os.path.join(self.grandparentdir, 'default', 'eventgen.conf'),
                            os.path.join(self.grandparentdir, 'local', 'eventgen.conf')]

            self.logger.debug('Reading configuration files for non-splunkembedded: %s' % conffiles)
            conf.read(conffiles)

            sections = conf.sections()
            ret = { }
            orig = { }
            for section in sections:
                ret[section] = dict(conf.items(section))
                # For compatibility with Splunk's configs, need to add the app name to an eai:acl key
                ret[section]['eai:acl'] = { 'app': self.grandparentdir.split(os.sep)[-1] }
            self._confDict = ret

        # Have to look in the data structure before normalization between what Splunk returns
        # versus what ConfigParser returns.
        logobj = logging.getLogger('eventgen')
        # if self._confDict['global']['debug'].lower() == 'true' \
        #         or self._confDict['global']['debug'].lower() == '1':
        #     logobj.setLevel(logging.DEBUG)
        # if self._confDict['global']['verbose'].lower() == 'true' \
        #         or self._confDict['global']['verbose'].lower() == '1':
        #     logobj.setLevel(logging.DEBUGV)

        # 2/1/15 CS  Adding support for command line options
        if self.args:
            if self.args.verbosity >= 2:
                self.debug = True
                logobj.setLevel(logging.DEBUG)
            if self.args.verbosity >= 3:
                self.verbose = True
                logobj.setLevel(logging.DEBUGV)
        self.logger.debug("ConfDict returned %s" % pprint.pformat(dict(self._confDict)))

def parse_args():
    """Parse command line arguments"""

    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("configfile", 
                        help="Location of eventgen.conf, app folder, or name of an app in $SPLUNK_HOME/etc/apps to run")
    parser.add_argument("-v", "--verbosity", action="count",
                        help="increase output verbosity")
    
    args = parser.parse_args()

    # Allow passing of a Splunk app on the command line and expand the full path before passing up the chain
    if not os.path.exists(args.configfile):
        if 'SPLUNK_HOME' in os.environ:
            if os.path.isdir(os.path.join(os.environ['SPLUNK_HOME'], 'etc', 'apps', args.configfile)):
                args.configfile = os.path.join(os.environ['SPLUNK_HOME'], 'etc', 'apps', args.configfile)
    return args

class Sample:
    """
    The Sample class is the primary configuration holder for Eventgen.  Contains all of our configuration
    information for any given sample, and is passed to most objects in Eventgen and a copy is maintained
    to give that object access to configuration information.  Read and configured at startup, and each
    object maintains a threadsafe copy of Sample.
    """
    # Required fields for Sample
    name = None
    app = None
    filePath = None
    
    # Options which are all valid for a sample
    disabled = None
    spoolDir = None
    spoolFile = None
    breaker = None
    sampletype = None
    mode = None
    interval = None
    delay = None
    count = None
    bundlelines = None
    earliest = None
    latest = None
    hourOfDayRate = None
    dayOfWeekRate = None
    randomizeEvents = None
    randomizeCount = None
    outputMode = None
    fileName = None
    fileMaxBytes = None
    fileBackupFiles = None
    splunkHost = None
    splunkPort = None
    splunkMethod = None
    splunkUser = None
    splunkPass = None
    index = None
    source = None
    sourcetype = None
    host = None
    hostRegex = None
    hostToken = None
    tokens = None
    projectID = None
    accessToken = None
    backfill = None
    backfillSearch = None
    backfillSearchUrl = None
    minuteOfHourRate = None
    timeMultiple = None
    debug = None
    timezone = datetime.timedelta(days=1)
    dayOfMonthRate = None
    monthOfYearRate = None
    sessionKey = None
    splunkUrl = None
    generator = None
    rater = None
    timeField = None
    timestamp = None
    sampleDir = None
    backfillts = None
    backfilldone = None
    stopping = False
    maxIntervalsBeforeFlush = None
    maxQueueLength = None
    end = None
    queueable = None
    autotimestamp = None

    
    # Internal fields
    sampleLines = None
    sampleDict = None
    _lockedSettings = None
    _priority = None
    _origName = None
    _lastts = None
    _earliestParsed = None
    _latestParsed = None
    
    def __init__(self, name):
        # 9/2/15 CS Can't make logger an attribute of the object like we do in other classes
        # because it borks deepcopy of the sample object
        logger = logging.getLogger('eventgen')
        adapter = EventgenAdapter(logger, {'module': 'Sample', 'sample': name})
        globals()['logger'] = adapter
        
        self.name = name
        self.tokens = [ ]
        self._lockedSettings = [ ]

        self.backfilldone = False
        
        # Import config
        globals()['c'] = Config()
        
    def __str__(self):
        """Only used for debugging, outputs a pretty printed representation of this sample"""
        filter_list = [ 'sampleLines', 'sampleDict' ]
        temp = dict([ (key, value) for (key, value) in self.__dict__.items() if key not in filter_list ])
        return pprint.pformat(temp)
        
    def __repr__(self):
        return self.__str__()
        
    ## Replaces $SPLUNK_HOME w/ correct pathing
    def pathParser(self, path):
        greatgreatgrandparentdir = os.path.dirname(os.path.dirname(c.grandparentdir)) 
        sharedStorage = ['$SPLUNK_HOME/etc/apps', '$SPLUNK_HOME/etc/users/', '$SPLUNK_HOME/var/run/splunk']

        ## Replace windows os.sep w/ nix os.sep
        path = path.replace('\\', '/')
        ## Normalize path to os.sep
        path = os.path.normpath(path)

        ## Iterate special paths
        for x in range(0, len(sharedStorage)):
            sharedPath = os.path.normpath(sharedStorage[x])

            if path.startswith(sharedPath):
                path.replace('$SPLUNK_HOME', greatgreatgrandparentdir)
                break

        ## Split path
        path = path.split(os.sep)

        ## Iterate path segments
        for x in range(0, len(path)):
            segment = path[x].lstrip('$')
            ## If segement is an environment variable then replace
            if os.environ.has_key(segment):
                path[x] = os.environ[segment]

        ## Join path
        path = os.sep.join(path)

        return path

    def _openSampleFile(self):
        logger.debugv("Opening sample '%s' in app '%s'" % (self.name, self.app))
        self._sampleFH = open(self.filePath, 'rU')

    def _closeSampleFile(self):
        logger.debugv("Closing sample '%s' in app '%s'" % (self.name, self.app))
        self._sampleFH.close()

    def loadSample(self):
        """Load sample from disk into self._sample.sampleLines and self._sample.sampleDict, 
        using cached copy if possible"""
        if self.sampletype == 'raw':
            # 5/27/12 CS Added caching of the sample file
            if self.sampleDict == None:
                self._openSampleFile()
                if self.breaker == c.breaker:
                    logger.debugv("Reading raw sample '%s' in app '%s'" % (self.name, self.app))
                    sampleLines = self._sampleFH.readlines()
                # 1/5/14 CS Moving to using only sampleDict and doing the breaking up into events at load time instead of on every generation
                else:
                    logger.debugv("Non-default breaker '%s' detected for sample '%s' in app '%s'" \
                                    % (self.breaker, self.name, self.app) ) 

                    sampleData = self._sampleFH.read()
                    sampleLines = [ ]

                    logger.debug("Filling array for sample '%s' in app '%s'; sampleData=%s, breaker=%s" \
                                    % (self.name, self.app, len(sampleData), self.breaker))

                    try:
                        breakerRE = re.compile(self.breaker, re.M)
                    except:
                        logger.error("Line breaker '%s' for sample '%s' in app '%s' could not be compiled; using default breaker" \
                                    % (self.breaker, self.name, self.app) )
                        self.breaker = c.breaker

                    # Loop through data, finding matches of the regular expression and breaking them up into
                    # "lines".  Each match includes the breaker itself.
                    extractpos = 0
                    searchpos = 0
                    breakerMatch = breakerRE.search(sampleData, searchpos)
                    while breakerMatch:
                        logger.debugv("Breaker found at: %d, %d" % (breakerMatch.span()[0], breakerMatch.span()[1]))
                        # Ignore matches at the beginning of the file
                        if breakerMatch.span()[0] != 0:
                            sampleLines.append(sampleData[extractpos:breakerMatch.span()[0]])
                            extractpos = breakerMatch.span()[0]
                        searchpos = breakerMatch.span()[1]
                        breakerMatch = breakerRE.search(sampleData, searchpos)
                    sampleLines.append(sampleData[extractpos:])

                self._closeSampleFile()

                self.sampleDict = [ { '_raw': line, 'index': self.index, 'host': self.host, 'source': self.source, 'sourcetype': self.sourcetype } for line in sampleLines ]
                logger.debug('Finished creating sampleDict & sampleLines.  Len samplesLines: %d Len sampleDict: %d' % (len(sampleLines), len(self.sampleDict)))
        elif self.sampletype == 'csv':
            if self.sampleDict == None:
                self._openSampleFile()
                logger.debugv("Reading csv sample '%s' in app '%s'" % (self.name, self.app))
                self.sampleDict = [ ]
                # Fix to load large csv files, work with python 2.5 onwards
                csv.field_size_limit(sys.maxint)
                csvReader = csv.DictReader(self._sampleFH)
                for line in csvReader:
                    if '_raw' in line:
                        self.sampleDict.append(line)
                    else:
                        logger.error("Missing _raw in line '%s'" % pprint.pformat(line))
                self._closeSampleFile()
                logger.debug("Finished creating sampleDict & sampleLines for sample '%s'.  Len sampleDict: %d" % (self.name, len(self.sampleDict)))

        # Ensure all lines have a newline
        for i in xrange(0, len(self.sampleDict)):
            if self.sampleDict[i]['_raw'][-1] != '\n':
                self.sampleDict[i]['_raw'] += '\n'

class Token:
    """Contains data and methods for replacing a token in a given sample"""
    token = None
    replacementType = None
    replacement = None
    sample = None
    mvhash = { }
    
    _replaytd = None
    _lastts = None
    _tokenfile = None
    _tokents = None
    _earliestTime = None
    _latestTime = None
    _replacementFile = None
    _replacementColumn = None
    _integerMatch = None
    _floatMatch = None
    _hexMatch = None
    _stringMatch = None
    _listMatch = None
    
    def __init__(self, sample=None):
        
        # Logger already setup by config, just get an instance
        logger = logging.getLogger('eventgen')
        if sample == None:
            name = "None"
        else:
            name = sample.name
        adapter = EventgenAdapter(logger, {'module': 'Token', 'sample': name})
        globals()['logger'] = adapter
        
        self._earliestTime = (None, None)
        self._latestTime = (None, None)
        
    def __str__(self):
        """Only used for debugging, outputs a pretty printed representation of this token"""
        # Eliminate recursive going back to parent
        temp = dict([ (key, value) for (key, value) in self.__dict__.items() if key != 'sample' ])
        return pprint.pformat(temp)

    def __repr__(self):
        return self.__str__()

if __name__ == '__main__':
    args = parse_args()
    c = Config(args)
    c.parse()

    export = { }
    export['global'] = {}
    export['global']['output'] = {}
    if c.fileName != None:
        export['global']['output']['fileName'] = c.fileName
    if c.fileBackupFiles != None:
        export['global']['output']['backupFiles'] = c.fileBackupFiles
    if c.fileMaxBytes != None:
        export['global']['output']['maxBytes'] = c.fileMaxBytes
    export['samples'] = []


    for s in c.samples:
        news = { }
        news['name'] = s.name
        news['begin'] = s.backfill
        news['count'] = s.count
        news['earliest'] = s.earliest
        news['latest'] = s.latest
        news['interval'] = s.interval
        news['randomizeCount'] = s.randomizeCount
        news['randomizeEvents'] = s.randomizeEvents
        news['lines'] = []
        for l in s.sampleDict:
            newline = { }
            for k, v in l.items():
                newline[k] = v.rstrip()
            news['lines'].append(newline)
        news['tokens'] = []
        group = 0
        groups = { }
        for i in xrange(0, len(s.tokens)):
            token = { }
            token['name'] = 'token.%d' % i
            token['format'] = 'regex'
            if re.match('.*\(.*\).*', s.tokens[i].token) == None:
                token['token'] = '(' + s.tokens[i].token + ')'
            else:
                token['token'] = s.tokens[i].token
            if s.tokens[i].replacementType == 'replaytimestamp':
                token['type'] = 'timestamp'
                token['replacement'] = s.tokens[i].replacement
            elif s.tokens[i].replacementType == 'file' or s.tokens[i].replacementType == 'mvfile':
                if s.tokens[i].replacement.find(':') > 0:
                    fname, col = s.tokens[i].replacement.split(':')
                else:
                    fname = s.tokens[i].replacement
                    col = 0
                token['name'] = os.path.basename(fname)
                token['sample'] = os.path.basename(fname)
                f = open(os.path.expandvars(fname), 'rU')
                if col > 0:
                    token['type'] = 'fieldChoice'
                    token['fieldChoice'] = []
                    reader = csv.reader(f)
                    for line in reader:
                        fc = { }
                        for j in xrange(0, len(line)):
                            fc[j+1] = line[j]
                        token['fieldChoice'].append(fc)
                    token['srcField'] = col
                    if fname not in groups:
                        group += 1
                        groups[fname] = group
                    token['group'] = groups[fname]
                else:
                    token['type'] = 'choice'
                    token['choice'] = []
                    for line in f:
                        token['choice'].append(line.rstrip())
            elif s.tokens[i].replacementType == 'random' or s.tokens[i].replacementType == 'rated':
                token['type'] = s.tokens[i].replacementType
                replacement = s.tokens[i].replacement

                integerRE = re.compile('integer\[([-]?\d+):([-]?\d+)\]', re.I)
                integerMatch = integerRE.match(replacement)
                floatRE = re.compile('float\[(\d+)\.(\d+):(\d+)\.(\d+)\]', re.I)
                floatMatch = floatRE.match(replacement)
                stringRE = re.compile('string\((\d+)\)', re.I)
                stringMatch = stringRE.match(replacement)
                hexRE = re.compile('hex\((\d+)\)', re.I)
                hexMatch = hexRE.match(replacement)
                listRE = re.compile('list(\[[^\]]+\])', re.I)
                listMatch = listRE.match(replacement)
                
                if integerMatch:
                    startInt = int(integerMatch.group(1))
                    endInt = int(integerMatch.group(2))
                    token['replacement'] = 'int'
                    token['lower'] = startInt
                    token['upper'] = endInt
                elif floatMatch:
                    startFloat = float(floatMatch.group(1)+'.'+floatMatch.group(2))
                    endFloat = float(floatMatch.group(3)+'.'+floatMatch.group(4))
                    token['replacement'] = 'float'
                    token['lower'] = startFloat
                    token['upper'] = endFloat
                    token['precision'] = len(startFloat.split('.')[1])
                elif stringMatch:
                    strLength = int(stringMatch.group(1))
                    token['replacement'] = 'string'
                    token['length'] = strLength
                elif hexMatch:
                    strLength = int(hexMatch.group(1))
                    token['replacement'] = 'hex'
                    token['length'] = strLength
                elif listMatch:
                    value = json.loads(listMatch.group(1))
                    token['type'] = 'choice'
                    token['choice'] = value
            else:
                token['type'] = s.tokens[i].replacementType
                token['replacement'] = s.tokens[i].replacement
                    
            news['tokens'].append(token)
        if s.fileName != None and 'fileName' not in export['global']['output']:
            export['global']['output']['fileName'] = s.fileName
        if s.fileBackupFiles != None and 'fileBackupFiles' not in export['global']['output']:
            export['global']['output']['backupFiles'] = s.fileBackupFiles
        if s.fileMaxBytes != None and 'fileMaxBytes' not in export['global']['output']:
            export['global']['output']['maxBytes'] = s.fileMaxBytes
        export['samples'].append(news)
        

    print json.dumps(export, indent=2)