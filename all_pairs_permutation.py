import re
import os
import itertools
from tqdm.auto import tqdm

print("Do not open any of the output files until the process is finished!")

print("Deleting old results...")
for file in tqdm(os.listdir()):
	if re.match(r'result_', file):
		os.remove(file)

with open ('.\\list_of_lists.txt', 'r', encoding = 'cp1251') as file:
    file = file.readlines()
    list_of_lists = [re.search(r'[A-Z].*\.xlsx', l).group(0) for l in file]

print("Permutation test for each pair of the given languages...")

for file1, file2 in tqdm(itertools.combinations(list_of_lists, 2)):
    print(f'Comparing {file1} and {file2}...')
    os.system(f'.\\bin\\spt_win_x86-64.exe --weights=.\\data\\weights.xlsx --set_a=.\\data\\{file1} --set_b=.\\data\\{file2} --output result.txt --consonants consonant.txt --cost_groups_plot=plot.svg')
