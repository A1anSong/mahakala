from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.cron import CronTrigger
import pytz

import analyze.chan as chan

scheduler = BackgroundScheduler()

minutes = '1'


def start_scheduler():
    # 周期为30分钟，每小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['30m'], trigger=CronTrigger(minute=f'{minutes}/30', timezone=pytz.UTC))
    # 周期为1小时，每小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['1h'], trigger=CronTrigger(minute=minutes, timezone=pytz.UTC))
    # 周期为2小时，每两小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['2h'], trigger=CronTrigger(hour='*/2', minute=minutes, timezone=pytz.UTC))
    # 周期为4小时，每四小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['4h'], trigger=CronTrigger(hour='*/4', minute=minutes, timezone=pytz.UTC))
    # 周期为6小时，每六小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['6h'], trigger=CronTrigger(hour='*/6', minute=minutes, timezone=pytz.UTC))
    # 周期为12小时，每十二小时的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['12h'], trigger=CronTrigger(hour='*/12', minute=minutes, timezone=pytz.UTC))
    # 周期为1天，每天的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['1d'],
                      trigger=CronTrigger(day='*', hour='0', minute=minutes, timezone=pytz.UTC))
    # 周期为3天，每三天的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['3d'],
                      trigger=CronTrigger(day='*/3', hour='0', minute=minutes, timezone=pytz.UTC))
    # 周期为5天，每五天的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['5d'],
                      trigger=CronTrigger(day='*/5', hour='0', minute=minutes, timezone=pytz.UTC))
    # 周期为1周，每周的第minutes分钟执行一次
    scheduler.add_job(chan.analyze, args=['1w'],
                      trigger=CronTrigger(day_of_week='1', hour='0', minute=minutes, timezone=pytz.UTC))

    scheduler.start()
