import numpy as np
import matplotlib.pyplot as plt
from matplotlib.animation import FuncAnimation
from matplotlib.animation import FFMpegWriter  # 動画保存用

# ファイルから座標データを読み込み
def read_coordinates(filename):
    with open(filename, 'r') as file:
        rounds = []
        current_round = []
        for line in file:
            line = line.strip()
            if line.startswith("Round"):
                if current_round:
                    # 現在のラウンドを追加する前に、ノード数が100未満の場合はパディングを追加
                    current_round = np.array(current_round, dtype=float)
                    if len(current_round) < 100:
                        padding = np.zeros((100 - len(current_round), current_round.shape[1]))
                        current_round = np.vstack((current_round, padding))
                    else:
                        current_round = current_round[:100]  # 100ノードに制限
                    rounds.append(current_round)
                    current_round = []
            else:
                coords = list(map(float, line.split()))
                current_round.append(coords)
        if current_round:  # 最後のラウンドを追加
            current_round = np.array(current_round, dtype=float)
            if len(current_round) < 100:
                padding = np.zeros((100 - len(current_round), current_round.shape[1]))
                current_round = np.vstack((current_round, padding))
            else:
                current_round = current_round[:100]  # 100ノードに制限
            rounds.append(current_round)

    return rounds

# アニメーションのセットアップ
def update_graph(num, data, scatter):
    x = data[num][:, 0]
    y = data[num][:, 1]
    z = data[num][:, 2]  # z 座標を取得
    scatter.set_offsets(np.c_[x, y])  # x, y 座標を設定
    scatter.set_array(z)  # z 座標を色として設定
    title.set_text(f"Round {num+1}")

# 座標データの読み込み
data = read_coordinates('vivaldi_node_position.txt')  # 入力ファイル名を指定

# アニメーションのプロット設定
fig, ax = plt.subplots()

# ノードの初期位置
initial_coords = data[0]
x = initial_coords[:, 0]
y = initial_coords[:, 1]
z = initial_coords[:, 2]  # z 座標

# 平面上に scatter プロットを作成 (色は z 座標で設定)
scatter = ax.scatter(x, y, c=z, cmap='viridis')

# カラーバーを追加して z 座標の範囲を示す
cbar = plt.colorbar(scatter, ax=ax)
cbar.set_label('Z coordinate (Color scale)')

# 軸の範囲設定
ax.set_xlim([-500, 500])
ax.set_ylim([-500, 500])

# タイトルを設定
title = ax.set_title("Round 1")

# アニメーション作成
ani = FuncAnimation(fig, update_graph, frames=len(data), fargs=(data, scatter), interval=500)

# 保存設定
writer = FFMpegWriter(fps=1)  # 1秒間隔でフレームを保存

# MP4ファイルとして保存
ani.save('vivaldi_node_position.mp4', writer=writer)

plt.show()
