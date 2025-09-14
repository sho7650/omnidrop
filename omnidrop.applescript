-- String utility functions
on splitString(inputString, delimiter)
    set oldDelimiters to AppleScript's text item delimiters
    set AppleScript's text item delimiters to delimiter
    set stringList to text items of inputString
    set AppleScript's text item delimiters to oldDelimiters
    return stringList
end splitString

on trimWhitespace(inputString)
    set trimmedString to inputString
    -- Remove leading whitespace
    repeat while trimmedString starts with " " or trimmedString starts with tab
        if length of trimmedString > 1 then
            set trimmedString to text 2 thru -1 of trimmedString
        else
            set trimmedString to ""
            exit repeat
        end if
    end repeat
    -- Remove trailing whitespace
    repeat while trimmedString ends with " " or trimmedString ends with tab
        if length of trimmedString > 1 then
            set trimmedString to text 1 thru -2 of trimmedString
        else
            set trimmedString to ""
            exit repeat
        end if
    end repeat
    return trimmedString
end trimWhitespace

-- Project resolution functions
on resolveProjectReference(projectPath, targetDocument)
    if projectPath is "" then
        return missing value
    end if
    
    try
        if projectPath contains "/" then
            return my resolveHierarchicalProject(projectPath, targetDocument)
        else
            return my resolveFlatProject(projectPath, targetDocument)
        end if
    on error errMsg
        error "Project resolution failed: " & errMsg
    end try
end resolveProjectReference

on resolveHierarchicalProject(projectPath, targetDocument)
    set pathComponents to my splitString(projectPath, "/")
    set projectName to last item of pathComponents
    
    -- If only one component, it's just a project name
    if (count of pathComponents) is 1 then
        return my resolveFlatProject(projectName, targetDocument)
    end if
    
    set folderComponents to items 1 thru -2 of pathComponents
    
    -- Build folder reference chain
    set currentContainer to targetDocument
    repeat with folderName in folderComponents
        set folderName to my trimWhitespace(folderName as string)
        try
            tell application "OmniFocus" to set folderList to (every folder of currentContainer whose name is folderName)
            if (count of folderList) > 0 then
                set currentContainer to item 1 of folderList
            else
                error "Folder not found: " & folderName
            end if
        on error errMsg
            error "Folder navigation failed: " & errMsg
        end try
    end repeat
    
    -- Find project in target folder
    try
        tell application "OmniFocus" to set projectList to (every project of currentContainer whose name is projectName)
        if (count of projectList) > 0 then
            return item 1 of projectList
        else
            error "Project '" & projectName & "' not found in folder hierarchy"
        end if
    on error errMsg
        error errMsg
    end try
end resolveHierarchicalProject

on resolveFlatProject(projectName, targetDocument)
    -- First try direct project search
    tell application "OmniFocus" to set projectList to (every project of targetDocument whose name is projectName)
    if (count of projectList) > 0 then
        return item 1 of projectList
    end if
    
    -- If not found, search in all folders recursively
    tell application "OmniFocus"
        tell targetDocument
            set allFolders to every folder
            repeat with aFolder in allFolders
                set foundProject to my searchProjectInFolder(projectName, aFolder)
                if foundProject is not missing value then
                    return foundProject
                end if
            end repeat
        end tell
    end tell
    
    error "Project not found: " & projectName
end resolveFlatProject

on searchProjectInFolder(projectName, targetFolder)
    tell application "OmniFocus"
        -- Check projects in this folder
        set projectList to (every project of targetFolder whose name is projectName)
        if (count of projectList) > 0 then
            return item 1 of projectList
        end if
        
        -- Check subfolders recursively
        set subFolders to every folder of targetFolder
        repeat with subFolder in subFolders
            set foundProject to my searchProjectInFolder(projectName, subFolder)
            if foundProject is not missing value then
                return foundProject
            end if
        end repeat
    end tell
    
    return missing value
end searchProjectInFolder

-- Main handler
on run argv
    -- Check arguments: expecting 4 arguments (title, note, project, tags)
    if (count of argv) < 4 then
        error "Expected 4 arguments: title, note, project, tags"
    end if

    -- Get arguments directly (no parsing needed)
    set taskTitle to item 1 of argv
    set taskNote to item 2 of argv
    set projectPath to item 3 of argv  -- Now supports hierarchical paths
    set tagsString to item 4 of argv

    try
        -- Validate title
        if taskTitle is "" then
            error "Title is required"
        end if

        -- Parse tags from comma-separated string
        set tagsList to {}
        if tagsString is not "" then
            set tagsList to my splitString(tagsString, ",")
        end if
        
        -- Create task in OmniFocus
        tell application "OmniFocus"
            tell default document
                -- Resolve project using new hierarchical system
                set targetProject to missing value
                if projectPath is not "" then
                    try
                        set docRef to it
                        set targetProject to my resolveProjectReference(projectPath, docRef)
                    on error errMsg
                        -- Log the error but continue (task will go to inbox)
                        log "Project resolution error: " & errMsg
                    end try
                end if
                
                -- Create task with proper project assignment
                if targetProject is not missing value then
                    tell targetProject
                        set newTask to make new task with properties {name:taskTitle}
                    end tell
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
                        set tagName to my trimWhitespace(tagName as text)
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