def identify_centers(df):
    # 上一个中枢的最高价和最低价
    last_center = (0, 0)
    # 上一个中枢的类型
    last_center_type = None

    # 在df中创建新的center列
    df['center_type_long'] = None
    df['center_type_short'] = None
    df['center_price'] = None

    # 过滤出有分型标记的数据
    df_fractal = df.dropna(subset=['fractal'])

    # 遍历有分型标记的数据
    for i in range(df_fractal.shape[0] - 4):
        # 如果第一个分型是底分型，那么就是上升中枢
        if df_fractal['fractal'].iloc[i] == 'bottom':
            current_low = min(df_fractal['Low'].iloc[i + 2], df_fractal['Low'].iloc[i + 4])
            current_high = max(df_fractal['High'].iloc[i + 1], df_fractal['High'].iloc[i + 3])
            # 如果第一个分型的底在当前中枢的高低之间，那么就不是有效的中枢
            if current_low <= df_fractal['Low'].iloc[i] <= current_high:
                continue
            # 如果上一个中枢也是上升中枢，那么判断这个中枢是否包含在上一个中枢中
            if last_center_type == 'long':
                if last_center[0] <= df_fractal['Low'].iloc[i + 2] <= last_center[1] \
                        or last_center[0] <= df_fractal['Low'].iloc[i + 4] <= last_center[1]:
                    continue
            # 中枢的顶是两个顶分型中最低的价格，中枢的底是两个底分型中最高的价格
            center_high = min(df_fractal['High'].iloc[i + 1], df_fractal['High'].iloc[i + 3])
            center_low = max(df_fractal['Low'].iloc[i + 2], df_fractal['Low'].iloc[i + 4])
            # 如果中枢的高点价格高于低点价格，那么中枢成立
            if center_low < center_high:
                df.loc[df_fractal.index[i + 1], 'center_type_long'] = 'start'
                df.loc[df_fractal.index[i + 1], 'center_price'] = center_high
                df.loc[df_fractal.index[i + 4], 'center_type_long'] = 'stop'
                df.loc[df_fractal.index[i + 4], 'center_price'] = center_low
                last_center = (current_low, current_high)
                last_center_type = 'long'
        # 如果第一个分型是顶分型，那么就是下降中枢
        if df_fractal['fractal'].iloc[i] == 'top':
            current_low = min(df_fractal['Low'].iloc[i + 1], df_fractal['Low'].iloc[i + 3])
            current_high = max(df_fractal['High'].iloc[i + 2], df_fractal['High'].iloc[i + 4])
            # 如果第一个分型的顶在当前中枢的高低之间，那么就不是有效的中枢
            if current_low <= df_fractal['High'].iloc[i] <= current_high:
                continue
            # 如果上一个中枢也是下降中枢，那么判断这个中枢是否包含在上一个中枢中
            if last_center_type == 'short':
                if last_center[0] <= df_fractal['High'].iloc[i + 2] <= last_center[1] \
                        or last_center[0] <= df_fractal['High'].iloc[i + 4] <= last_center[1]:
                    continue
            # 中枢的顶是两个顶分型中最低的价格，中枢的底是两个底分型中最高的价格
            center_high = min(df_fractal['High'].iloc[i + 2], df_fractal['High'].iloc[i + 4])
            center_low = max(df_fractal['Low'].iloc[i + 1], df_fractal['Low'].iloc[i + 3])
            # 如果中枢的高点价格高于低点价格，那么中枢成立
            if center_low < center_high:
                df.loc[df_fractal.index[i + 1], 'center_type_short'] = 'start'
                df.loc[df_fractal.index[i + 1], 'center_price'] = center_low
                df.loc[df_fractal.index[i + 4], 'center_type_short'] = 'stop'
                df.loc[df_fractal.index[i + 4], 'center_price'] = center_high
                last_center = (current_low, current_high)
                last_center_type = 'short'

    return df
