
#!/bin/zsh

# Ensure script stops on error
set -e

# Prompt to delete existing 'slickchat' tasks for a clean setup
read -q "?Deseja apagar todas as tarefas 'slickchat' existentes antes de configurá-las? (y/N) "
echo
if [[ "$REPLY" == "y" || "$REPLY" == "Y" ]]; then
    echo "Apagando tarefas 'slickchat' existentes..."
    # '|| true' allows the command to fail if no tasks are found, without stopping the script
    task project:slickchat delete --force || true 
    echo "Tarefas 'slickchat' existentes foram apagadas."
fi

declare -A task_id_map # Associative array to map markdown order to actual Taskwarrior IDs
local task_counter=0
local -a task_lines_from_md=() # Store the raw 'task add' lines from .task.md

# First pass: Read all 'task add' commands from .task.md
while IFS= read -r line; do
    # Only process lines that start with 'task add' and are not comments or empty
    if [[ "$line" =~ ^task\ add\ .* ]]; then
        task_lines_from_md+=("$line")
    fi
done < .task.md

# Second pass: Process and add tasks sequentially
for task_cmd_original in "${task_lines_from_md[@]}"; do
    task_counter=$((task_counter + 1))
    local task_cmd_base="$task_cmd_original"
    local depends_resolved_str=""

    # Check for 'depends:' in the command line
    # The regex now captures the optional preceding space as match[1] and the depends clause as match[2]
    if [[ "$task_cmd_base" =~ ([[:space:]]+)?(depends:[0-9,]+) ]]; then
        local depends_full_match="${match[0]}" # e.g., " depends:1,2,3" or "depends:1,2,3"
        local dep_numbers_str="${match[2]#depends:}" # e.g., "1,2,3"

        # Remove the original depends clause (including its preceding space if any) from the command base
        # This prevents duplicating 'depends' and ensures we add the resolved ones
        task_cmd_base="${task_cmd_base//$depends_full_match/}"
        # Clean up any potential double spaces left by the removal
        task_cmd_base="${task_cmd_base//  / }"
        # Remove trailing space if the depends clause was at the end and had a leading space
        task_cmd_base="${task_cmd_base/%[[:space:]]/}"

        local -a dep_numbers_array
        # Split dependency numbers by comma into an array
        IFS=',' read -r -A dep_numbers_array <<< "$dep_numbers_str"

        local -a resolved_deps_ids=()
        for dep_num in "${dep_numbers_array[@]}"; do
            if [[ -n "$dep_num" && -v "task_id_map[$dep_num]" ]]; then
                # If the dependency number (from markdown) is resolved, add its Taskwarrior ID
                resolved_deps_ids+=("${task_id_map[$dep_num]}")
            else
                echo "Aviso: Não foi possível resolver a dependência para a tarefa markdown #$dep_num (dependência da tarefa atual #$task_counter). Pulando esta dependência."
            fi
        done

        if (( ${#resolved_deps_ids[@]} > 0 )); then
            # Join resolved Taskwarrior IDs with commas for the 'depends:' clause
            depends_resolved_str="depends:${(j:,:)resolved_deps_ids}"
        fi
    fi

    # Construct the final command to execute
    local final_task_cmd="$task_cmd_base"
    if [[ -n "$depends_resolved_str" ]]; then
        final_task_cmd+=" $depends_resolved_str"
    fi

    echo "Adicionando tarefa markdown #$task_counter: $final_task_cmd"
    
    # Execute the command and capture its output
    local output
    # Using eval is necessary here because the command string is constructed dynamically.
    # Given that the source is .task.md (a controlled file), this is acceptable.
    output=$(eval "$final_task_cmd")
    echo "$output"

    # Extract the Taskwarrior ID from the output (e.g., "Created task 1." or "Created task 1 (UUID:...)")
    if [[ "$output" =~ "Created task "([0-9]+) ]]; then
        local created_id="${match[1]}"
        task_id_map[$task_counter]="$created_id"
        echo "Mapeada tarefa markdown #$task_counter para Taskwarrior ID: $created_id"
    else
        echo "Erro: Não foi possível extrair o ID do Taskwarrior para a tarefa markdown #$task_counter. Saída: $output"
        exit 1 # Exit if a task ID cannot be extracted, as this could break future dependencies
    fi
    echo "" # Add a newline for better readability between tasks
done

echo "Todas as tarefas de .task.md foram processadas com sucesso."
echo "\nMapeamento da ordem da tarefa no markdown para os IDs do Taskwarrior:"
for k v in "${(@kv)task_id_map}"; do
    echo "  Tarefa Markdown #$k -> Taskwarrior ID $v"
done

echo "\nPara visualizar suas novas tarefas, execute: task project:slickchat"
