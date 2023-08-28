def add_macd(df):
    # 计算快速移动平均线
    df['Fast EMA'] = df['Close'].ewm(span=12, adjust=False).mean()
    # 计算慢速移动平均线
    df['Slow EMA'] = df['Close'].ewm(span=26, adjust=False).mean()
    # 计算离差值
    df['DIF'] = df['Fast EMA'] - df['Slow EMA']
    # 计算离差平均值
    df['DEA'] = df['DIF'].ewm(span=9, adjust=False).mean()
    # 计算MACD柱状图
    df['MACD'] = 2 * (df['DIF'] - df['DEA'])
    return df
