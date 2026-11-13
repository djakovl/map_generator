import csv
import numpy as np
import matplotlib.pyplot as plt
from matplotlib.colors import ListedColormap

# Цвета из classify.go (нормализованные от 0 до 1)
colors = [
    [28/255, 88/255, 176/255],   # 0: Вода (Синий)
    [222/255, 202/255, 122/255], # 1: Пляж (Желтый)
    [118/255, 188/255, 92/255],  # 2: Равнина (Светло-зеленый)
    [42/255, 122/255, 62/255],   # 3: Лес (Темно-зеленый)
    [232/255, 212/255, 134/255], # 4: Пустыня
    [128/255, 128/255, 128/255], # 5: Горы
    [242/255, 246/255, 250/255]  # 6: Снег
]
cmap = ListedColormap(colors)

def load_csv(path):
    with open(path, 'r') as f:
        reader = csv.reader(f)
        return np.array([[float(v) for v in row] for row in reader])

# Загружаем данные классификации
terrain = load_csv('results/terrain.csv')

plt.figure(figsize=(8, 8))
plt.imshow(terrain, cmap=cmap, vmin=0, vmax=6)
plt.title('Карта биомов (Классификация)')

# ИСПРАВЛЕНИЕ: создаем colorbar и задаем метки отдельным методом
cbar = plt.colorbar(ticks=[0, 1, 2, 3, 4, 5, 6])
cbar.set_ticklabels(['Вода', 'Пляж', 'Равнина', 'Лес', 'Пустыня', 'Горы', 'Снег'])

plt.tight_layout()
plt.savefig('biome_map.png', dpi=150)
plt.show()
print("✅ Карта биомов успешно сохранена в biome_map.png")