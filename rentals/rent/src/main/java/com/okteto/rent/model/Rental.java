package com.okteto.rent.model;

import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import jakarta.persistence.Column;

@Entity
@Table(name = "rentals")
public class Rental {

    @Id
    @Column(name = "id", nullable = false, unique = true)
    private String id;

    @Column(name = "price", nullable = false)
    private String price;

    public Rental() {
    }

    public Rental(String id, String price) {
        this.id = id;
        this.price = price;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getPrice() {
        return price;
    }

    public void setPrice(String price) {
        this.price = price;
    }
}
