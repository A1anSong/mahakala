def merge_candle(df):
    drop_rows = []
    i = 0
    last_index = 0
    final_keep_index = 0
    while i < df.shape[0] - 1:
        j = i + 1
        if last_index == final_keep_index:
            last_index = i - 1
            final_keep_index = i - 1
        else:
            last_index = final_keep_index
        curr_row = df.iloc[i]
        next_row = df.iloc[j]
        while i > 0 and ((curr_row['High'] >= next_row['High'] and curr_row['Low'] <= next_row['Low']) or (
                curr_row['High'] <= next_row['High'] and curr_row['Low'] >= next_row['Low'])):
            keep_index = i
            drop_index = j
            last_row = df.iloc[last_index]
            # 如果当前K线被下一根K线包含，那么就删除当前K线
            if curr_row['High'] <= next_row['High'] and curr_row['Low'] >= next_row['Low']:
                keep_index = j
                drop_index = i
            # 如果是上升
            if curr_row['High'] >= last_row['High']:
                df.loc[df.index[keep_index], 'High'] = max(curr_row['High'], next_row['High'])
                df.loc[df.index[keep_index], 'Low'] = max(curr_row['Low'], next_row['Low'])
                df.loc[df.index[keep_index], 'Open'] = df.loc[df.index[keep_index], 'Low']
                df.loc[df.index[keep_index], 'Close'] = df.loc[df.index[keep_index], 'High']
            # 如果是下降
            else:
                df.loc[df.index[keep_index], 'High'] = min(curr_row['High'], next_row['High'])
                df.loc[df.index[keep_index], 'Low'] = min(curr_row['Low'], next_row['Low'])
                df.loc[df.index[keep_index], 'Open'] = df.loc[df.index[keep_index], 'High']
                df.loc[df.index[keep_index], 'Close'] = df.loc[df.index[keep_index], 'Low']
            df.loc[df.index[keep_index], 'Volume'] = curr_row['Volume'] + next_row['Volume']
            df.loc[df.index[keep_index], 'DIF'] = curr_row['DIF'] + next_row['DIF']
            df.loc[df.index[keep_index], 'DEA'] = curr_row['DEA'] + next_row['DEA']
            df.loc[df.index[keep_index], 'MACD'] = curr_row['MACD'] + next_row['MACD']
            final_keep_index = keep_index
            drop_rows.append(df.index[drop_index])
            if j < df.shape[0] - 1:
                j += 1
                if drop_index == i:
                    i = keep_index
                curr_row = df.iloc[i]
                next_row = df.iloc[j]
            else:
                break
        i = j
    df = df.drop(drop_rows)
    return df
