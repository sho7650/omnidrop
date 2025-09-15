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
                
                -- Set due date to today at 18:00:00
                set todayDate to current date
                set hours of todayDate to 18
                set minutes of todayDate to 00
                set seconds of todayDate to 00
                set due date of newTask to todayDate
                
                -- Set tags using multi-strategy approach with fallbacks
                if (count of tagsList) > 0 then
                    set tagResults to my assignTagsWithFallback(newTask, tagsList, it)
                    set assignedTags to assigned of tagResults
                    set failedTags to failed of tagResults
                    
                    -- Log summary results
                    if (count of assignedTags) > 0 then
                        log "Task created with " & (count of assignedTags) & " tags: " & my listToString(assignedTags)
                    end if
                    if (count of failedTags) > 0 then
                        log "Warning: " & (count of failedTags) & " tags could not be assigned: " & my listToString(failedTags)
                    end if
                end if
            end tell
        end tell
        
        return "success"
        
    on error errMsg
        error errMsg
    end try
end run

-- Helper function: Get or create tag with safe context handling
on getOrCreateTagSafely(tagName, docRef)
    tell application "OmniFocus"
        tell docRef
            try
                set tagList to (every tag whose name is tagName)
                if (count of tagList) > 0 then
                    log "Found existing tag: " & tagName
                    return item 1 of tagList
                else
                    set newTag to make new tag with properties {name:tagName}
                    log "Created new tag: " & tagName
                    return newTag
                end if
            on error errMsg
                log "Error in getOrCreateTagSafely for '" & tagName & "': " & errMsg
                return missing value
            end try
        end tell
    end tell
end getOrCreateTagSafely

-- Helper function: Multi-strategy tag assignment with fallbacks
on assignTagsWithFallback(newTask, tagsList, docRef)
    set assignedTags to {}
    set failedTags to {}
    
    repeat with tagName in tagsList
        set tagName to my trimWhitespace(tagName as text)
        if tagName is not "" then
            set tagRef to my getOrCreateTagSafely(tagName, docRef)
            set tagAssigned to false
            
            if tagRef is not missing value then
                -- Strategy A: Individual tag assignment
                try
                    tell application "OmniFocus"
                        tell docRef
                            add tagRef to tags of newTask
                        end tell
                    end tell
                    set end of assignedTags to tagName
                    set tagAssigned to true
                    log "Strategy A success - Individual assignment for: " & tagName
                on error errMsg1
                    log "Strategy A failed for '" & tagName & "': " & errMsg1

                    -- Strategy B: Direct tag property assignment
                    try
                        tell application "OmniFocus"
                            tell docRef
                                tell newTask
                                    set tags to tags & {tagRef}
                                end tell
                            end tell
                        end tell
                        set end of assignedTags to tagName
                        set tagAssigned to true
                        log "Strategy B success - Property assignment for: " & tagName
                    on error errMsg2
                        log "Strategy B failed for '" & tagName & "': " & errMsg2
                        set end of failedTags to tagName
                    end try
                end try
            else
                set end of failedTags to tagName
            end if
        end if
    end repeat
    
    -- Strategy C: If no tags assigned and we have tags, try primary tag approach
    if (count of assignedTags) = 0 and (count of tagsList) > 0 then
        set firstTagName to my trimWhitespace(item 1 of tagsList as text)
        if firstTagName is not "" then
            try
                set primaryTagRef to my getOrCreateTagSafely(firstTagName, docRef)
                if primaryTagRef is not missing value then
                    tell application "OmniFocus"
                        tell docRef
                            set primary tag of newTask to primaryTagRef
                        end tell
                    end tell
                    set end of assignedTags to firstTagName
                    log "Strategy C success - Primary tag assignment for: " & firstTagName
                end if
            on error errMsg3
                log "Strategy C failed for '" & firstTagName & "': " & errMsg3
            end try
        end if
    end if
    
    -- Log final results
    if (count of assignedTags) > 0 then
        log "Successfully assigned tags: " & my listToString(assignedTags)
    end if
    if (count of failedTags) > 0 then
        log "Failed to assign tags: " & my listToString(failedTags)
    end if
    
    return {assigned:assignedTags, failed:failedTags}
end assignTagsWithFallback

-- Helper function: Convert list to comma-separated string
on listToString(theList)
    set AppleScript's text item delimiters to ", "
    set theString to theList as string
    set AppleScript's text item delimiters to ""
    return theString
end listToString