#!/usr/bin/env python3
"""
Go项目依赖关系分析工具 - 分析模块间依赖，找出核心模块和学习路径
"""

import os
import re
from collections import defaultdict, Counter
from dataclasses import dataclass
from typing import Set, Dict, List, Tuple


@dataclass
class ModuleInfo:
    name: str
    path: str
    imports: Set[str]
    internal_imports: Set[str]  # 项目内部导入
    external_imports: Set[str]  # 外部库导入
    exported_functions: Set[str]
    file_count: int
    code_lines: int


class DependencyAnalyzer:
    def __init__(self, root_dir='.'):
        self.root_dir = root_dir
        self.project_module_name = self.detect_module_name()
        
    def detect_module_name(self) -> str:
        """从go.mod检测项目模块名"""
        go_mod_path = os.path.join(self.root_dir, 'go.mod')
        if os.path.exists(go_mod_path):
            try:
                with open(go_mod_path, 'r') as f:
                    for line in f:
                        if line.startswith('module '):
                            return line.split()[1].strip()
            except:
                pass
        return "unknown"
    
    def is_internal_import(self, import_path: str) -> bool:
        """判断是否是项目内部导入"""
        return import_path.startswith(self.project_module_name)
    
    def extract_imports(self, file_path: str) -> Tuple[Set[str], Set[str]]:
        """提取文件的导入信息"""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
        except:
            return set(), set()
        
        imports = set()
        
        # 匹配单行导入
        single_imports = re.findall(r'import\s+"([^"]+)"', content)
        imports.update(single_imports)
        
        # 匹配多行导入块
        import_blocks = re.findall(r'import\s*\((.*?)\)', content, re.DOTALL)
        for block in import_blocks:
            block_imports = re.findall(r'"([^"]+)"', block)
            imports.update(block_imports)
        
        # 分离内部和外部导入
        internal = {imp for imp in imports if self.is_internal_import(imp)}
        external = imports - internal
        
        return internal, external
    
    def extract_exported_functions(self, file_path: str) -> Set[str]:
        """提取导出的函数名（大写开头）"""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
        except:
            return set()
        
        # 匹配导出函数（大写开头）
        functions = re.findall(r'func\s+(?:\([^)]*\)\s+)?([A-Z]\w*)\s*\(', content)
        return set(functions)
    
    def count_code_lines(self, file_path: str) -> int:
        """统计代码行数（排除注释和空行）"""
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                lines = f.readlines()
        except:
            return 0
        
        code_lines = 0
        for line in lines:
            stripped = line.strip()
            if stripped and not stripped.startswith('//'):
                code_lines += 1
        return code_lines
    
    def get_module_from_path(self, file_path: str) -> str:
        """从文件路径推导模块名"""
        rel_path = os.path.relpath(file_path, self.root_dir)
        parts = rel_path.split(os.sep)[:-1]  # 去掉文件名
        
        # 过滤掉一些不重要的路径部分
        filtered_parts = []
        for part in parts:
            if part not in {'.', '..', 'internal'}:
                filtered_parts.append(part)
        
        return '/'.join(filtered_parts) if filtered_parts else 'root'
    
    def analyze_modules(self) -> Dict[str, ModuleInfo]:
        """分析所有模块"""
        modules = defaultdict(lambda: {
            'imports': set(),
            'internal_imports': set(),
            'external_imports': set(),
            'exported_functions': set(),
            'files': [],
            'code_lines': 0
        })
        
        # 遍历所有Go文件
        for root, dirs, files in os.walk(self.root_dir):
            # 跳过一些目录
            dirs[:] = [d for d in dirs if not d.startswith('.') and d not in ['vendor', 'node_modules']]
            
            for file in files:
                if file.endswith('.go') and not file.endswith('_test.go'):
                    file_path = os.path.join(root, file)
                    module_name = self.get_module_from_path(file_path)
                    
                    internal_imports, external_imports = self.extract_imports(file_path)
                    exported_functions = self.extract_exported_functions(file_path)
                    code_lines = self.count_code_lines(file_path)
                    
                    modules[module_name]['imports'].update(internal_imports | external_imports)
                    modules[module_name]['internal_imports'].update(internal_imports)
                    modules[module_name]['external_imports'].update(external_imports)
                    modules[module_name]['exported_functions'].update(exported_functions)
                    modules[module_name]['files'].append(file_path)
                    modules[module_name]['code_lines'] += code_lines
        
        # 转换为ModuleInfo对象
        result = {}
        for name, data in modules.items():
            if data['files']:  # 只包含有文件的模块
                result[name] = ModuleInfo(
                    name=name,
                    path=os.path.dirname(data['files'][0]) if data['files'] else '',
                    imports=data['imports'],
                    internal_imports=data['internal_imports'],
                    external_imports=data['external_imports'],
                    exported_functions=data['exported_functions'],
                    file_count=len(data['files']),
                    code_lines=data['code_lines']
                )
        
        return result
    
    def calculate_module_importance(self, modules: Dict[str, ModuleInfo]) -> Dict[str, float]:
        """计算模块重要性评分"""
        scores = {}
        
        # 统计每个模块被多少其他模块依赖
        dependency_count = defaultdict(int)
        for module in modules.values():
            for imp in module.internal_imports:
                # 将导入路径转换为模块名
                imp_module = imp.replace(self.project_module_name + '/', '')
                dependency_count[imp_module] += 1
        
        for name, module in modules.items():
            score = 0
            
            # 被依赖次数（越多越重要）
            score += dependency_count.get(name, 0) * 10
            
            # 导出函数数量（越多越重要）
            score += len(module.exported_functions) * 2
            
            # 代码行数（适中最好，太多太少都不是核心）
            lines = module.code_lines
            if 100 <= lines <= 1000:
                score += 5
            elif 50 <= lines <= 2000:
                score += 3
            
            # 特殊路径加分
            if any(keyword in name.lower() for keyword in ['api', 'types', 'core', 'common']):
                score += 15
            
            # CLI和测试代码减分
            if any(keyword in name.lower() for keyword in ['cmd', 'test', 'example']):
                score -= 5
            
            scores[name] = score
        
        return scores
    
    def find_learning_path(self, modules: Dict[str, ModuleInfo], importance_scores: Dict[str, float]) -> List[str]:
        """推荐学习路径"""
        # 按重要性和依赖关系排序
        def learning_priority(module_name):
            module = modules[module_name]
            importance = importance_scores.get(module_name, 0)
            complexity = len(module.internal_imports)  # 依赖越少越容易理解
            
            # 优先学习：重要且简单的模块
            return importance - complexity
        
        sorted_modules = sorted(modules.keys(), key=learning_priority, reverse=True)
        
        # 过滤掉一些不重要的模块
        learning_path = []
        for module_name in sorted_modules:
            module = modules[module_name]
            if (module.code_lines > 20 and  # 至少有一定代码量
                not any(skip in module_name.lower() for skip in ['test', 'example', 'mock']) and
                len(learning_path) < 15):  # 最多推荐15个模块
                learning_path.append(module_name)
        
        return learning_path
    
    def print_analysis(self, modules: Dict[str, ModuleInfo]):
        """打印分析结果"""
        print("="*80)
        print("🔗 Go项目依赖关系分析")
        print("="*80)
        
        importance_scores = self.calculate_module_importance(modules)
        
        print(f"\n📊 项目概览:")
        print(f"  • 项目模块名: {self.project_module_name}")
        print(f"  • 总模块数: {len(modules)}")
        print(f"  • 总代码行数: {sum(m.code_lines for m in modules.values()):,}")
        
        # 按重要性排序显示模块
        sorted_by_importance = sorted(modules.items(), 
                                    key=lambda x: importance_scores.get(x[0], 0), 
                                    reverse=True)
        
        print(f"\n📈 模块重要性排名 (Top 10):")
        print("-" * 70)
        print(f"{'模块名':<30} {'重要性':>8} {'代码行':>8} {'导出函数':>8} {'依赖':>6}")
        print("-" * 70)
        
        for i, (name, module) in enumerate(sorted_by_importance[:10]):
            score = importance_scores.get(name, 0)
            print(f"{name:<30} {score:>8.1f} {module.code_lines:>8} {len(module.exported_functions):>8} {len(module.internal_imports):>6}")
        
        # 推荐学习路径
        learning_path = self.find_learning_path(modules, importance_scores)
        
        print(f"\n🎯 推荐学习路径:")
        print("-" * 50)
        
        for i, module_name in enumerate(learning_path[:8], 1):
            module = modules[module_name]
            score = importance_scores.get(module_name, 0)
            
            # 推荐原因
            reasons = []
            if score > 20:
                reasons.append("核心模块")
            if len(module.exported_functions) > 5:
                reasons.append("接口丰富")
            if len(module.internal_imports) <= 3:
                reasons.append("依赖简单")
            if any(keyword in module_name for keyword in ['api', 'types']):
                reasons.append("数据定义")
            
            reason_str = ", ".join(reasons[:2]) if reasons else "基础模块"
            
            print(f"{i:2d}. {module_name}")
            print(f"    📄 {module.file_count}个文件, {module.code_lines}行代码")
            print(f"    🔧 {len(module.exported_functions)}个导出函数")
            print(f"    💡 {reason_str}")
            print()
        
        # 外部依赖统计
        all_external = set()
        for module in modules.values():
            all_external.update(module.external_imports)
        
        # 统计最常用的外部库
        external_usage = Counter()
        for module in modules.values():
            for ext in module.external_imports:
                external_usage[ext] += 1
        
        print(f"\n📦 主要外部依赖 (Top 10):")
        print("-" * 50)
        for lib, count in external_usage.most_common(10):
            print(f"  • {lib} (被{count}个模块使用)")


def main():
    analyzer = DependencyAnalyzer()
    modules = analyzer.analyze_modules()
    analyzer.print_analysis(modules)


if __name__ == "__main__":
    main() 