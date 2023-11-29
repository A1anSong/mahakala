def identify_fractal(df):
    """
    识别顶分型和底分型
    """

    # 创建一个新的列来存储分型
    df['fractal'] = None

    # 识别顶分型
    df.loc[(df['High'].shift(1) < df['High']) &
           (df['High'].shift(-1) < df['High']), 'fractal'] = 'top'

    # 识别底分型
    df.loc[(df['Low'].shift(1) > df['Low']) &
           (df['Low'].shift(-1) > df['Low']), 'fractal'] = 'bottom'

    return df
