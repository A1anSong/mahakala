import mplfinance as mpf
import pandas as pd
import io


def draw_klines(df, symbol, interval):
    addplot_all = add_plots(df)
    all_lines = add_lines(df)
    rectangles = add_rectangles(df)
    buf = io.BytesIO()
    fig_scale = (len(df) + 1) / 100
    # 绘制图表
    if len(rectangles) > 0:
        mpf.plot(df, figscale=fig_scale, type='candle', style='binance', title=f'{symbol} {interval}',
                 ylabel='Price (₮)', volume=True, ylabel_lower='Volume', volume_panel=2, addplot=addplot_all,
                 alines=all_lines, fill_between=rectangles, warn_too_much_data=1000, savefig=buf)
    else:
        mpf.plot(df, figscale=fig_scale, type='candle', style='binance', title=f'{symbol} {interval}',
                 ylabel='Price (₮)', volume=True, ylabel_lower='Volume', volume_panel=2, addplot=addplot_all,
                 alines=all_lines, warn_too_much_data=1000, savefig=buf)
    buf.seek(0)

    return buf


# 绘制中枢
def add_rectangles(df):
    # 过滤出'center_type_long'和'center_type_short'列不为空的行
    df_centered_long = df.loc[df['center_type_long'].notna()]
    df_centered_short = df.loc[df['center_type_short'].notna()]

    # 初始化一个空的矩形列表
    rectangles_long = []
    rectangles_short = []

    # 遍历df_centered_long中的所有行，找到所有的中枢
    start_time = None
    for index, row in df_centered_long.iterrows():
        if row['center_type_long'] == 'start':
            start_time = index
            y1 = row['center_price']
        elif row['center_type_long'] == 'stop':
            if start_time is not None:
                stop_time = index
                where_values = (df.index >= start_time) & (df.index <= stop_time)
                rectangle = dict(y1=y1, y2=row['center_price'], where=where_values, alpha=0.4, color='g')
                rectangles_long.append(rectangle)

    # 遍历df_centered_short中的所有行，找到所有的中枢
    start_time = None
    for index, row in df_centered_short.iterrows():
        if row['center_type_short'] == 'start':
            start_time = index
            y1 = row['center_price']
        elif row['center_type_short'] == 'stop':
            if start_time is not None:
                stop_time = index
                where_values = (df.index >= start_time) & (df.index <= stop_time)
                rectangle = dict(y1=y1, y2=row['center_price'], where=where_values, alpha=0.4, color='r')
                rectangles_short.append(rectangle)

    rectangles = rectangles_long + rectangles_short

    return rectangles


# 绘制线段
def add_lines(df):
    # 创建一个新的DataFrame，只包含有分型的行
    df_fractals = df.dropna(subset=['fractal'])
    # 初始化一个空列表用于存储分型和对应的价格
    fractals_lines = []
    # 在df_centered中遍历所有有分型的数据
    for idx, row in df_fractals.iterrows():
        # 根据分型类型选择价格
        price = row['High'] if row['fractal'] == 'top' else row['Low']

        # 将日期和价格组成一个元组，并添加到列表中
        fractals_lines.append((idx, price))

    all_lines = dict(alines=fractals_lines, colors='c', linewidths=0.5)

    return all_lines


# 绘制附图
def add_plots(df):
    # 创建布林带和 MACD 的附图
    ap_mid_band = mpf.make_addplot(df['Middle Band'], panel=0, color='orange')  # 将布林带设为面板0
    ap_upper_band = mpf.make_addplot(df['Upper Band'], panel=0, color='red')
    ap_lower_band = mpf.make_addplot(df['Lower Band'], panel=0, color='blue')
    ap_dif = mpf.make_addplot(df['DIF'], panel=1, color='b', secondary_y=False)  # 将MACD设为面板1
    ap_dea = mpf.make_addplot(df['DEA'], panel=1, color='y', secondary_y=False)
    ap_macd = mpf.make_addplot(df['MACD'], panel=1, color='dimgray', secondary_y=False, type='bar')
    # 创建两个布尔数组，用于标记顶分型和底分型
    tops = (df['fractal'] == 'top')
    bottoms = (df['fractal'] == 'bottom')
    # 创建两个新的Series，长度与df_identified相同
    tops_series = pd.Series(index=df.index)
    bottoms_series = pd.Series(index=df.index)
    # 对于顶分型和底分型，将价格填入相应的Series
    tops_series[tops] = df['High'][tops]
    bottoms_series[bottoms] = df['Low'][bottoms]
    # 使用make_addplot()来创建额外的绘图，用于标记顶分型和底分型
    addplot_tops = mpf.make_addplot(tops_series, scatter=True, markersize=50, marker='v', color='r')
    addplot_bottoms = mpf.make_addplot(bottoms_series, scatter=True, markersize=50, marker='^', color='g')

    addplot_all = [ap_mid_band, ap_upper_band, ap_lower_band, ap_dif, ap_dea, ap_macd, addplot_tops, addplot_bottoms]

    return addplot_all
