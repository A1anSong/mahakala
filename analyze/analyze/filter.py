def filter_fractals(df):
    # 设置一个标记来跟踪最后一个有效的分型是顶分型还是底分型
    last_valid_fractal = None
    last_valid_fractal_index = None
    # 再设置一个标记来跟踪倒数第二个有效的分型是顶分型还是底分型
    pre_last_valid_fractal = None
    pre_last_valid_fractal_index = None

    # 找出所有的分型
    fractals = df.loc[df['fractal'].notnull()].copy()

    # 创建shift列
    fractals['next_row'] = df.index.to_series().shift(-1)
    fractals['prev_row'] = df.index.to_series().shift(1)

    for index, row in fractals.iterrows():
        # 如果还没有找到任何有效的分型，那么当前的分型就是有效的
        if last_valid_fractal is None:
            last_valid_fractal = row
            last_valid_fractal_index = index
        # 检查当前分型是否满足有效性规则
        else:
            # 如果当前分型和上一个分型是同一类型的
            if row['fractal'] == last_valid_fractal['fractal']:
                # 新的顶分型的高点比之前有效的顶分型的高点还要高
                if row['fractal'] == 'top':
                    if row['High'] > last_valid_fractal['High']:
                        df.loc[last_valid_fractal_index, 'fractal'] = None
                        last_valid_fractal = row
                        last_valid_fractal_index = index
                    else:
                        df.loc[index, 'fractal'] = None
                # 新的底分型的低点比之前有效底分型的低点还要低
                if row['fractal'] == 'bottom':
                    if row['Low'] < last_valid_fractal['Low']:
                        df.loc[last_valid_fractal_index, 'fractal'] = None
                        last_valid_fractal = row
                        last_valid_fractal_index = index
                    else:
                        df.loc[index, 'fractal'] = None
            # 顶分型的最高点必须高于前一个底分型的最低点
            # 底分型的最低点必须低于前一个顶分型的最高点
            elif ((row['fractal'] == 'top' and row['High'] > last_valid_fractal['Low']) or
                  (row['fractal'] == 'bottom' and row['Low'] < last_valid_fractal['High'])):
                # 两个有效分型之间必须有至少一根K线
                if df.loc[row['prev_row'], 'index'] - df.loc[last_valid_fractal['next_row'], 'index'] > 1:
                    pre_last_valid_fractal, last_valid_fractal = last_valid_fractal, row
                    pre_last_valid_fractal_index, last_valid_fractal_index = last_valid_fractal_index, index
                else:
                    if pre_last_valid_fractal is not None:
                        if row['fractal'] == 'top':
                            if row['High'] > pre_last_valid_fractal['High']:
                                df.loc[pre_last_valid_fractal_index, 'fractal'] = None
                                df.loc[last_valid_fractal_index, 'fractal'] = None
                                last_valid_fractal = row
                                last_valid_fractal_index = index
                                pre_last_valid_fractal = None
                                pre_last_valid_fractal_index = None
                            else:
                                df.loc[index, 'fractal'] = None
                        if row['fractal'] == 'bottom':
                            if row['Low'] < pre_last_valid_fractal['Low']:
                                df.loc[pre_last_valid_fractal_index, 'fractal'] = None
                                df.loc[last_valid_fractal_index, 'fractal'] = None
                                last_valid_fractal = row
                                last_valid_fractal_index = index
                                pre_last_valid_fractal = None
                                pre_last_valid_fractal_index = None
                            else:
                                df.loc[index, 'fractal'] = None
                    else:
                        df.loc[index, 'fractal'] = None
            else:
                df.loc[index, 'fractal'] = None

    return df
