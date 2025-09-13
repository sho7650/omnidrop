on run argv
    -- Check arguments: expecting 4 arguments (title, note, project, tags)
    if (count of argv) < 4 then
        error "Expected 4 arguments: title, note, project, tags"
    end if

    -- Get arguments directly (no parsing needed)
    set taskTitle to item 1 of argv
    set taskNote to item 2 of argv
    set projectName to item 3 of argv
    set tagsString to item 4 of argv

    try
        -- Validate title
        if taskTitle is "" then
            error "Title is required"
        end if

        -- Parse tags from comma-separated string
        set tagsList to {}
        if tagsString is not "" then
            set oldDelimiters to AppleScript's text item delimiters
            set AppleScript's text item delimiters to ","
            set tagsList to text items of tagsString
            set AppleScript's text item delimiters to oldDelimiters
        end if
        
        -- Create task in OmniFocus
        tell application "OmniFocus"
            tell default document
                -- Find project if specified
                set targetProject to missing value
                if projectName is not "" then
                    try
                        set projectList to (every project whose name is projectName)
                        if (count of projectList) > 0 then
                            set targetProject to item 1 of projectList
                        end if
                    end try
                end if
                
                -- Create task
                if targetProject is not missing value then
                    set newTask to make new task with properties {name:taskTitle}
                    move newTask to end of tasks of targetProject
                else
                    set newTask to make new inbox task with properties {name:taskTitle}
                end if
                
                -- Set note if provided
                if taskNote is not "" then
                    set note of newTask to taskNote
                end if
                
                -- Set due date to today at 23:59:59
                set todayDate to current date
                set hours of todayDate to 23
                set minutes of todayDate to 59
                set seconds of todayDate to 59
                set due date of newTask to todayDate
                
                -- Set tags if provided
                if (count of tagsList) > 0 then
                    set resolvedTags to {}
                    repeat with tagName in tagsList
                        set tagName to tagName as text
                        if tagName is not "" then
                            try
                                set tagList to (every tag whose name is tagName)
                                if (count of tagList) > 0 then
                                    set end of resolvedTags to item 1 of tagList
                                end if
                            end try
                        end if
                    end repeat
                    if (count of resolvedTags) > 0 then
                        set tags of newTask to resolvedTags
                    end if
                end if
            end tell
        end tell
        
        return "success"
        
    on error errMsg
        error errMsg
    end try
end run