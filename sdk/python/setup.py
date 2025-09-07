from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

with open("requirements.txt", "r", encoding="utf-8") as fh:
    requirements = [line.strip() for line in fh if line.strip() and not line.startswith("#")]

setup(
    name="adrenochain-sdk",
    version="1.0.0",
    author="Adrenochain Team",
    author_email="dev@adrenochain.com",
    description="Python SDK for Adrenochain blockchain network",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/palaseus/adrenochain",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Software Development :: Libraries :: Python Modules",
        "Topic :: Office/Business :: Financial :: Investment",
    ],
    python_requires=">=3.8",
    install_requires=requirements,
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-cov>=4.0.0",
            "black>=22.0.0",
            "flake8>=5.0.0",
            "mypy>=1.0.0",
            "pre-commit>=2.20.0",
        ],
        "jupyter": [
            "jupyter>=1.0.0",
            "ipywidgets>=8.0.0",
            "matplotlib>=3.5.0",
            "seaborn>=0.11.0",
            "plotly>=5.0.0",
        ],
        "trading": [
            "pandas>=1.5.0",
            "numpy>=1.21.0",
            "scikit-learn>=1.1.0",
            "ta-lib>=0.4.0",
            "ccxt>=3.0.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "adrenochain-cli=adrenochain.cli:main",
        ],
    },
    include_package_data=True,
    package_data={
        "adrenochain": ["py.typed"],
    },
)
