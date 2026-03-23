package data

type Metadata struct {
    CurrentPage int `json:"current_page,omitempty"`
    PageSize int `json:"page_size,omitempty"`
    FirstPage int `json:"first_page,omitempty"`
    LastPage int `json:"last_page,omitempty"`
    TotalRecords int `json:"total_records,omitempty"`
}

// Calculate the metadata
func calculateMetaData(totalRecords int, currentPage int, pageSize int) Metadata {
    if totalRecords == 0 {
        return Metadata{}
    }

    return Metadata {
        CurrentPage: currentPage,
        PageSize: pageSize,
        FirstPage: 1,
        LastPage: (totalRecords + pageSize - 1) / pageSize,
        TotalRecords: totalRecords,
   }
    
}