# crystal

Crystal is designed to make certain kinds of automation easier and more robust.
It does this by flowing data using files as nodes and programs as edges. 
The Crystal program listens for files being modified and automatically calls the associated script.

The main configuration file is crystalfile, where a regex for files is declared, along with node and output file.
The format for crystalfile is as follows:

`input_regex node_name output_file`
