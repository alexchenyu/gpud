#!/usr/bin/env python3
"""
Go代码行数统计工具
统计当前项目中所有.go文件的代码行数
"""

import os
import pathlib
from collections import defaultdict


def count_lines_in_file(file_path):
    """统计单个文件的行数"""
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            lines = f.readlines()
        
        total_lines = len(lines)
        empty_lines = sum(1 for line in lines if line.strip() == '')
        comment_lines = sum(1 for line in lines if line.strip().startswith('//'))
        code_lines = total_lines - empty_lines - comment_lines
        
        return {
            'total': total_lines,
            'code': code_lines,
            'comments': comment_lines,
            'empty': empty_lines
        }
    except Exception as e:
        print(f"警告: 无法读取文件 {file_path}: {e}")
        return {'total': 0, 'code': 0, 'comments': 0, 'empty': 0}


def find_go_files(root_dir):
    """递归查找所有.go文件"""
    go_files = []
    for root, dirs, files in os.walk(root_dir):
        # 跳过常见的不需要统计的目录
        dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ['vendor', 'node_modules']]
        
        for file in files:
            if file.endswith('.go'):
                go_files.append(os.path.join(root, file))
    
    return sorted(go_files)


def main():
    """主函数"""
    root_dir = '.'
    
    print("🔍 正在扫描Go文件...")
    go_files = find_go_files(root_dir)
    
    if not go_files:
        print("❌ 未找到任何Go文件")
        return
    
    print(f"📁 找到 {len(go_files)} 个Go文件")
    print()
    
    # 按目录分组统计
    dir_stats = defaultdict(lambda: {'files': 0, 'total': 0, 'code': 0, 'comments': 0, 'empty': 0})
    total_stats = {'files': 0, 'total': 0, 'code': 0, 'comments': 0, 'empty': 0}
    
    print("📊 详细统计:")
    print("-" * 80)
    print(f"{'文件路径':<50} {'总行数':>8} {'代码行':>8} {'注释行':>8} {'空行':>6}")
    print("-" * 80)
    
    for file_path in go_files:
        stats = count_lines_in_file(file_path)
        
        # 相对路径显示
        rel_path = os.path.relpath(file_path, root_dir)
        
        # 输出文件统计
        print(f"{rel_path:<50} {stats['total']:>8} {stats['code']:>8} {stats['comments']:>8} {stats['empty']:>6}")
        
        # 目录统计
        dir_name = os.path.dirname(rel_path) or '.'
        dir_stats[dir_name]['files'] += 1
        for key in ['total', 'code', 'comments', 'empty']:
            dir_stats[dir_name][key] += stats[key]
        
        # 总计统计
        total_stats['files'] += 1
        for key in ['total', 'code', 'comments', 'empty']:
            total_stats[key] += stats[key]
    
    print("-" * 80)
    print()
    
    # 按目录汇总
    print("📂 按目录统计:")
    print("-" * 70)
    print(f"{'目录':<30} {'文件数':>8} {'总行数':>8} {'代码行':>8} {'注释行':>8} {'空行':>6}")
    print("-" * 70)
    
    for dir_name in sorted(dir_stats.keys()):
        stats = dir_stats[dir_name]
        print(f"{dir_name:<30} {stats['files']:>8} {stats['total']:>8} {stats['code']:>8} {stats['comments']:>8} {stats['empty']:>6}")
    
    print("-" * 70)
    print()
    
    # 总计
    print("📈 总计统计:")
    print("=" * 50)
    print(f"Go文件总数:    {total_stats['files']:,}")
    print(f"总行数:        {total_stats['total']:,}")
    print(f"代码行数:      {total_stats['code']:,}")
    print(f"注释行数:      {total_stats['comments']:,}")
    print(f"空行数:        {total_stats['empty']:,}")
    print("=" * 50)
    
    # 计算比例
    if total_stats['total'] > 0:
        code_percentage = (total_stats['code'] / total_stats['total']) * 100
        comment_percentage = (total_stats['comments'] / total_stats['total']) * 100
        empty_percentage = (total_stats['empty'] / total_stats['total']) * 100
        
        print()
        print("📊 代码构成比例:")
        print(f"代码行:   {code_percentage:.1f}%")
        print(f"注释行:   {comment_percentage:.1f}%")
        print(f"空行:     {empty_percentage:.1f}%")
    
    # 找出最大的文件
    if go_files:
        file_sizes = [(f, count_lines_in_file(f)['total']) for f in go_files]
        largest_files = sorted(file_sizes, key=lambda x: x[1], reverse=True)[:5]
        
        print()
        print("🏆 代码行数最多的5个文件:")
        for i, (file_path, lines) in enumerate(largest_files, 1):
            rel_path = os.path.relpath(file_path, root_dir)
            print(f"{i}. {rel_path}: {lines:,} 行")


if __name__ == "__main__":
    main() 