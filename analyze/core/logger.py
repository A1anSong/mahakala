import logging.config

import core.config as core_config

config = core_config.config

logging.config.dictConfig(config['logging'])
logger = logging.getLogger('console_logger')
logger.propagate = False
