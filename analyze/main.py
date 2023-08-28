import utils.scheduler as scheduler
import server.flask as flask

web = flask.web

if __name__ == '__main__':
    # 启动定时任务
    scheduler.start_scheduler()
    # 启动Web服务
    web.run()
