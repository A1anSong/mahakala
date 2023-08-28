import yaml
import os.path

from watchdog.events import FileSystemEventHandler
from watchdog.observers import Observer

with open('config.yaml', 'r') as f:
    config = yaml.safe_load(f)


class ConfigFileHandler(FileSystemEventHandler):
    def on_modified(self, event):
        if os.path.basename(event.src_path) == 'config.yaml':
            with open('config.yaml', 'r') as f:
                global config
                config = yaml.safe_load(f)
            print('config.yaml has been modified')
            print(config['mahakala'])


file_handler = ConfigFileHandler()
observer = Observer()
observer.schedule(file_handler, './', recursive=False)
observer.start()
